package aichat

import (
	"regexp"
	"slices"
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
var draftInjuryClauseBreak = regexp.MustCompile(`[, ]+(?:but|and) +(?:, *)?(i\b|we\b|not sure\b|unsure\b)|,\s+(no (?:current |active |known )?(?:injur(?:y|ies)|(?:movement )?limitations)\b)`)
var draftSelfClearance = regexp.MustCompile(`^(?:i (?:also |currently )?have no |i (?:don't|do not) have |i am (?:injury|pain)[ -]free\b|i'm (?:injury|pain)[ -]free\b|no |(?:injury|pain)[ -]free\b)`)
var draftSelfReport = regexp.MustCompile(`^(?:i(?: |'m |'ve )|my (?:right |left )?(?:ankle|arm|back|elbow|foot|hand|hip|knee|leg|neck|shoulder|wrist)\b)`)
var draftOtherInjurySubject = regexp.MustCompile(`\b(?:my (?:training |workout )?(?:partner|friend|coach|client|sibling|parent|mom|dad|wife|husband)|(?:he|she|they) (?:has|have|had|is|are|hurt)|for (?:him|her|them))\b`)
var draftMoreInjuries = regexp.MustCompile(`\b(?:any (?:other|additional)|anything else|what else)\b`)
var draftLimitedNegation = regexp.MustCompile(`\bno (?:sharp |severe |shooting |acute |new )+(?:pain|injur(?:y|ies))\b`)

func draftReportedLimitation(text string) bool {
	return draftRestriction.MatchString(text) ||
		draftInjuryTerms.MatchString(draftLimitedNegation.ReplaceAllString(draftNoInjury.ReplaceAllString(text, ""), ""))
}

// Injury clearance comes from user messages or a recorded profile.
// A model may point to a report, but only the enclosing user-authored
// clause can establish its subject, polarity, and current applicability.
func (c draftInjuryContext) confirmedStatus(evidence string) string {
	quote := normalizeQualityText(evidence)
	var restrictions []string
	var cleared bool
	if c.profile != nil && c.profile.MovementLimitationsRecorded {
		cleared = len(c.profile.MovementLimitations) == 0
		restrictions = append(restrictions, c.profile.MovementLimitations...)
	}

	record := func(message string, preceding *RuntimeChatMessage) {
		text := normalizeQualityText(message)
		question := ""
		if preceding != nil && preceding.Role == roleAssistant {
			question = normalizeQualityText(preceding.Text)
		}
		answering := strings.Contains(question, "?") && draftInjuryTerms.MatchString(question)
		moreInjuries := draftMoreInjuries.MatchString(question)
		if answering && !moreInjuries && len(restrictions) == 0 {
			switch strings.Trim(text, " .!?") {
			case "no", "none", "nope", "nothing":
				cleared = true
				return
			}
		}

		// Split separate statements without severing uncertainty from its
		// object (e.g. "I'm not sure whether I have any injuries").
		text = draftInjuryClauseBreak.ReplaceAllString(text, "; $1$2")
		var reported []string
		var fullClearance, uncertain bool
		for _, clause := range strings.FieldsFunc(text, func(r rune) bool { return strings.ContainsRune(".!?;", r) }) {
			clause = strings.Trim(clause, " ,")
			if draftOtherInjurySubject.MatchString(clause) {
				continue
			}
			if draftClearanceWithheld.MatchString(clause) &&
				(draftInjuryTerms.MatchString(clause) || draftRestriction.MatchString(clause) ||
					(answering && (clause == "not sure" || clause == "i'm not sure" || clause == "unsure"))) {
				uncertain = true
				continue
			}
			if draftReportedLimitation(clause) &&
				(draftSelfReport.MatchString(clause) || (answering && (hasAnyQualityTerm(clause, "it's", "it is", "recent", "mild", "moderate") ||
					draftRestriction.MatchString(clause) || (quote != "" && strings.Contains(clause, quote))))) {
				reported = append(reported, clause)
				continue
			}
			if !moreInjuries && draftSelfClearance.MatchString(clause) && draftNoInjury.MatchString(clause) {
				fullClearance = true
			}
		}
		if fullClearance && !uncertain {
			restrictions, cleared = nil, true
		}
		if uncertain {
			cleared = false
		}
		for _, limitation := range reported {
			if !slices.Contains(restrictions, limitation) {
				restrictions = append(restrictions, limitation)
			}
		}
	}

	var preceding *RuntimeChatMessage
	for i := range c.history {
		message := &c.history[i]
		if message.Role == "user" {
			record(message.Text, preceding)
		}
		preceding = message
	}
	record(c.prompt, preceding)
	if len(restrictions) > 0 {
		return strings.Join(restrictions, "; ")
	}
	if cleared {
		return "none"
	}
	return ""
}
