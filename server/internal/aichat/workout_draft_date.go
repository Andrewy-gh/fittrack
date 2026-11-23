package aichat

import (
	"context"
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

func resolveWorkoutDraftSchedule(requested string, now time.Time) (workoutDraftSchedule, bool) {
	requested = strings.ToLower(strings.TrimSpace(requested))
	if requested == "" || requested == "today" {
		return workoutDraftSchedule{date: now, historyThrough: now}, true
	}
	if exact, err := time.Parse(time.RFC3339, strings.ToUpper(requested)); err == nil {
		return workoutDraftSchedule{date: exact, historyThrough: earlierWorkoutDraftTime(exact, now)}, true
	}
	var day time.Time
	switch requested {
	case "tomorrow":
		day = now.AddDate(0, 0, 1)
	case "yesterday":
		day = now.AddDate(0, 0, -1)
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
