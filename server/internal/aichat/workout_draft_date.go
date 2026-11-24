package aichat

import (
	"context"
	"regexp"
	"strings"
	"time"
	_ "time/tzdata" // The production image does not ship the system timezone database.
)

type draftTimezoneKey struct{}
type draftClockKey struct{}

func withDraftTimezone(ctx context.Context, zone string) context.Context {
	if zone == "" || zone == "Local" || len(zone) > 100 {
		return ctx
	}
	if location, err := time.LoadLocation(zone); err == nil {
		return context.WithValue(ctx, draftTimezoneKey{}, location)
	}
	return ctx
}

func draftTimezone(ctx context.Context) *time.Location {
	if location, ok := ctx.Value(draftTimezoneKey{}).(*time.Location); ok {
		return location
	}
	return time.UTC
}

func draftClock(ctx context.Context) time.Time {
	now, _ := ctx.Value(draftClockKey{}).(time.Time)
	if now.IsZero() {
		now = time.Now()
	}
	return now.In(draftTimezone(ctx))
}

type workoutDraftSchedule struct {
	date           time.Time
	historyThrough time.Time
}

var workoutRelativeDay = regexp.MustCompile(`\b(?:for|on|schedule|reschedule|move|workout|work out|train|training|session) (?:it (?:to|for) )?(today|tomorrow|yesterday)\b|\b(today|tomorrow|yesterday)['’]s (?:(?:upper[- ]body|lower[- ]body|full[- ]body|push|pull|leg|legs|chest|back|strength|cardio|mobility|\d+[- ]minute) ){0,3}(?:workout|training|session)\b`)

// With a relative tool argument, anchor it to the most recent user turn that
// supplied that wording rather than a follow-up answer about another detail.
func relativeWorkoutDraftAnchor(requested string, source draftInjuryContext, now time.Time) (time.Time, bool) {
	requested = normalizeQualityText(requested)
	if requested != "today" && requested != "tomorrow" && requested != "yesterday" {
		return now, true
	}
	mentionsRequestedDay := func(text string) bool {
		text = normalizeQualityText(text)
		if strings.Trim(text, " ,.?!") == requested {
			return true
		}
		// Merely mentioning "today" in a status answer doesn't reschedule.
		for _, match := range workoutRelativeDay.FindAllStringSubmatch(text, -1) {
			if match[1] == requested || match[2] == requested {
				return true
			}
		}
		return false
	}
	if mentionsRequestedDay(source.prompt) {
		return now, true
	}
	for i := len(source.history) - 1; i >= 0; i-- {
		message := source.history[i]
		if message.Role == roleUser && mentionsRequestedDay(message.Text) {
			if message.CreatedAt.IsZero() {
				break
			}
			return message.CreatedAt.In(now.Location()), true
		}
	}
	// An ungrounded relative date must be clarified, not interpreted using
	// the timestamp of an unrelated follow-up.
	return time.Time{}, false
}

func resolveWorkoutDraftSchedule(requested string, now, anchor time.Time) (workoutDraftSchedule, bool) {
	requested = strings.ToLower(strings.TrimSpace(requested))
	if requested == "" {
		return workoutDraftSchedule{date: now, historyThrough: now}, true
	}
	if exact, err := time.Parse(time.RFC3339, strings.ToUpper(requested)); err == nil {
		return workoutDraftSchedule{date: exact, historyThrough: earlierWorkoutDraftTime(exact, now)}, true
	}
	var day time.Time
	if anchor.IsZero() {
		anchor = now
	}
	switch requested {
	case "today":
		day = anchor
	case "tomorrow":
		day = anchor.AddDate(0, 0, 1)
	case "yesterday":
		day = anchor.AddDate(0, 0, -1)
	default:
		var err error
		day, err = time.ParseInLocation(time.DateOnly, requested, now.Location())
		if err != nil {
			return workoutDraftSchedule{}, false
		}
	}
	year, month, date := day.Date()
	if y, m, d := now.Date(); year == y && month == m && date == d {
		return workoutDraftSchedule{date: now, historyThrough: now}, true
	}
	localNoon := time.Date(year, month, date, 12, 0, 0, 0, now.Location())
	// A date with no time includes history from the whole requested local day.
	localEnd := time.Date(year, month, date+1, 0, 0, 0, 0, now.Location()).Add(-time.Microsecond)
	return workoutDraftSchedule{date: localNoon, historyThrough: earlierWorkoutDraftTime(localEnd, now)}, true
}

func earlierWorkoutDraftTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}
