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
var draftInjuryClauseBreak = regexp.MustCompile(`[, ]+(?:but|and) +(?:, *)?(i\b|we\b|not sure\b|unsure\b)|,\s+(no (?:current |active |known )?(?:injur(?:y|ies)|(?:movement )?limitations)\b|i\b|my\b|we\b|not sure\b|unsure\b)`)
var draftSelfClearance = regexp.MustCompile(`^(?:i (?:also |currently )?have no |i (?:don't|do not) have |i am (?:injury|pain)[ -]free\b|i'm (?:injury|pain)[ -]free\b|no |(?:injury|pain)[ -]free\b)`)
var draftSelfReport = regexp.MustCompile(`^(?:i (?:also |still |currently |recently )?(?:have|feel|hurt|strained|sprained|can't|cannot|need to avoid|must avoid)\b|i'm (?:having|dealing with|unable to)\b|i've (?:had|hurt|strained|sprained)\b)`)
var draftOwnBody = regexp.MustCompile(`^my (?:right |left )?(?:ankles?|arms?|back|elbows?|feet|foot|hands?|hips?|knees?|legs?|neck|shoulders?|wrists?)\b`)
var draftOtherInjurySubject = regexp.MustCompile(`\b(?:my (?:training |workout )?(?:partner|friend|coach|client|sibling|parents?|mom|dad|wife|husband|daughter|son|child|kids?)\b|(?:he|she|they) (?:has|have|had|is|are|hurt)|for (?:him|her|them))\b`)

const draftInjuryQuestion = "Do you have any current injuries or movement limitations?"

// Bare answers must follow a standalone question directed to the user.
// Don't infer who "No" refers to in arbitrary assistant prose.
func isDraftInjuryQuestion(text string) bool {
	switch text {
	case "do you have any current injuries or movement limitations?", "do you have any current injuries?", "do you have any injuries?", "any current injuries?", "any injuries?":
		return true
	}
	return false
}

var draftMoreInjuries = regexp.MustCompile(`\b(?:any (?:other|additional)|anything else|what else)\b`)
var draftLimitedNegation = regexp.MustCompile(`\bno (?:sharp |severe |shooting |acute |new )+(?:pain|injur(?:y|ies))\b`)

func isShortInjuryNegative(text string) bool {
	switch strings.TrimSpace(text) {
	case "no", "none", "nope", "nothing":
		return true
	}
	return false
}

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
		answering := isDraftInjuryQuestion(question)
		moreInjuries := draftMoreInjuries.MatchString(question)

		// Split separate statements without severing uncertainty from its
		// object (e.g. "I'm not sure whether I have any injuries").
		text = draftInjuryClauseBreak.ReplaceAllString(text, "; $1$2")
		var reported []string
		var fullClearance, uncertain bool
		if prefix, _, hasComma := strings.Cut(text, ","); answering && len(restrictions) == 0 && hasComma && isShortInjuryNegative(prefix) {
			fullClearance = true
		}
		for i, clause := range strings.FieldsFunc(text, func(r rune) bool { return strings.ContainsRune(".!?;", r) }) {
			clause = strings.Trim(clause, " ,")
			if i == 0 && answering && len(restrictions) == 0 {
				if isShortInjuryNegative(clause) {
					fullClearance = true
					continue
				}
			}
			ownBody := draftOwnBody.MatchString(clause)
			if !ownBody && draftOtherInjurySubject.MatchString(clause) {
				continue
			}
			if draftClearanceWithheld.MatchString(clause) &&
				(draftInjuryTerms.MatchString(clause) || draftRestriction.MatchString(clause) ||
					(answering && (clause == "not sure" || clause == "i'm not sure" || clause == "unsure"))) {
				uncertain = true
				continue
			}
			if draftReportedLimitation(clause) &&
				(ownBody || draftSelfReport.MatchString(clause) || answering ||
					(quote != "" && strings.Contains(clause, quote) && draftRestriction.MatchString(clause))) {
				reported = append(reported, clause)
				continue
			}
			// A terse “no injuries” after a noncanonical assistant injury
			// question has an uncertain subject. Ask for a first-person answer.
			ambiguousSubject := !answering && strings.Contains(question, "?") && draftInjuryTerms.MatchString(question) &&
				!strings.HasPrefix(clause, "i ")
			if !moreInjuries && !ambiguousSubject && draftSelfClearance.MatchString(clause) && draftNoInjury.MatchString(clause) {
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
