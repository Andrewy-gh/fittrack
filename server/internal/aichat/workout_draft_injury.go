package aichat

import (
	"regexp"
	"strings"
)

type draftInjuryContextKey struct{}

type draftInjuryContext struct {
	prompt  string
	history []RuntimeChatMessage
	profile *TrainingProfile
}

var draftNoInjury = regexp.MustCompile(`\b(?:no (?:current |active |known )?(?:injur(?:y|ies)|pain|(?:movement )?limitations)(?:(?:, | and | or )(?:injur(?:y|ies)|pain|(?:movement )?limitations))*|(?:i |we )?(?:don't|do not) have (?:any )?(?:current |active )?(?:injur(?:y|ies)|pain|(?:movement )?limitations)|(?:injury|pain)[ -]free)\b`)
var draftInjuryTerms = regexp.MustCompile(`\b(?:injur\w*|pain\w*|hurt\w*|sore\w*|stiff\w*|strain\w*|sprain\w*|discomfort|limitations?)\b`)
var draftRestriction = regexp.MustCompile(`\b(?:can't|cannot|unable to|need to avoid|must avoid) (?:(?:fully|safely|comfortably|easily|completely|currently) )*(?:raise|lift|bend|move|press|reach|grip|squat|kneel|run|jump|overhead|heavy)\b`)
var draftClearanceWithheld = regexp.MustCompile(`\b(?:not sure|not confirmed|unconfirmed|unsure|uncertain|(?:can't|cannot|won't|don't|do not) (?:confirm|know|say|assume|claim)|(?:haven't|have not|hasn't|has not) (?:confirmed|established|verified)|(?:if|whether) (?:i|we) (?:have|had))\b`)
var draftLimitedNegation = regexp.MustCompile(`\bno (?:sharp |severe |shooting |acute |new )+(?:pain|injur(?:y|ies))\b`)

func draftReportedLimitation(text string) bool {
	return draftRestriction.MatchString(text) ||
		draftInjuryTerms.MatchString(draftLimitedNegation.ReplaceAllString(draftNoInjury.ReplaceAllString(text, ""), ""))
}

// Injury clearance comes from user messages or a recorded profile. The model
// can identify a reported limitation, but its quote must occur in a user message.
func (c draftInjuryContext) confirmedStatus(evidence string) string {
	quote := normalizeQualityText(evidence)
	check := func(message string, preceding *RuntimeChatMessage) (string, bool) {
		text := normalizeQualityText(message)
		answering := preceding != nil && preceding.Role == "assistant" &&
			strings.Contains(preceding.Text, "?") && draftInjuryTerms.MatchString(normalizeQualityText(preceding.Text))
		var restrictions []string
		clearance := false
		clauses := strings.FieldsFunc(strings.ReplaceAll(text, " but ", ";"), func(r rune) bool {
			return strings.ContainsRune(".!?;", r)
		})
		for _, clause := range clauses {
			clause = strings.Trim(clause, " ,")
			if (draftInjuryTerms.MatchString(clause) || draftRestriction.MatchString(clause)) && draftClearanceWithheld.MatchString(clause) {
				return "", true
			}
			if draftReportedLimitation(clause) &&
				(answering || hasAnyQualityTerm(clause, "i have", "my", "it's", "it is", "recent", "mild", "moderate", "from my") || draftRestriction.MatchString(clause)) {
				restrictions = append(restrictions, clause)
			}
			clearance = clearance || draftNoInjury.MatchString(clause)
		}
		if len(restrictions) > 0 {
			return strings.Join(restrictions, "; "), true
		}
		if quote != "" && strings.Contains(text, quote) && draftReportedLimitation(quote) && !draftClearanceWithheld.MatchString(quote) {
			return quote, true
		}
		if draftReportedLimitation(text) {
			return "", true
		}
		if clearance {
			return "none", true
		}
		if answering {
			switch strings.Trim(text, " .!?") {
			case "no", "none", "nope", "nothing":
				return "none", true
			}
		}
		return "", false
	}
	previous := func(index int) *RuntimeChatMessage {
		if index > 0 {
			return &c.history[index-1]
		}
		return nil
	}
	if status, found := check(c.prompt, previous(len(c.history))); found {
		return status
	}
	for i := len(c.history) - 1; i >= 0; i-- {
		if c.history[i].Role == "user" {
			if status, found := check(c.history[i].Text, previous(i)); found {
				return status
			}
		}
	}
	if c.profile != nil && c.profile.MovementLimitationsRecorded {
		if len(c.profile.MovementLimitations) == 0 {
			return "none"
		}
		return strings.Join(c.profile.MovementLimitations, ", ")
	}
	return ""
}
