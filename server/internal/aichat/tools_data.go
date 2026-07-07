package aichat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const (
	getWorkoutsToolName        = "get_workouts"
	getWorkoutsToolDescription = "Reads the authenticated user's logged FitTrack workouts. Use this for questions about their personal workout history, recent sessions, exercises performed, or logged training data. Do not use it for general fitness knowledge or to create a new workout draft."
)

type GetWorkoutsToolInput struct {
	LastN        int    `json:"lastN,omitempty" jsonschema:"description=How many recent workouts (max 20). Default 5."`
	StartDate    string `json:"startDate,omitempty" jsonschema:"description=Inclusive ISO date (YYYY-MM-DD). Resolve relative phrases yourself using the current date."`
	EndDate      string `json:"endDate,omitempty" jsonschema:"description=Inclusive ISO date (YYYY-MM-DD). Resolve relative phrases yourself using the current date."`
	ExerciseName string `json:"exerciseName,omitempty" jsonschema:"description=Partial name ok, e.g. 'bench'."`
	WorkoutFocus string `json:"workoutFocus,omitempty" jsonschema:"description=Optional workout focus filter such as push, pull, legs, upper body, lower body, or full body."`
}

type GetWorkoutsToolResult struct {
	Workouts           []ChatWorkoutView `json:"workouts,omitempty"`
	Message            string            `json:"message,omitempty"`
	CandidateExercises []string          `json:"candidateExercises,omitempty"`
}

func defineGetWorkoutsTool(g *genkit.Genkit, reader ChatDataReader) ai.Tool {
	return genkit.DefineTool(g, getWorkoutsToolName,
		getWorkoutsToolDescription,
		func(ctx *ai.ToolContext, input GetWorkoutsToolInput) (*GetWorkoutsToolResult, error) {
			startedAt := time.Now()
			logAIChatTraceContext(ctx, "get_workouts_tool_started",
				"last_n", input.LastN,
				"exercise_name", strings.TrimSpace(input.ExerciseName),
				"workout_focus", strings.TrimSpace(input.WorkoutFocus),
				"request_id", request.GetRequestID(ctx),
			)
			result := runGetWorkoutsTool(ctx, reader, input)
			logAIChatTraceContext(ctx, "get_workouts_tool_finished",
				"elapsed_ms", time.Since(startedAt).Milliseconds(),
				"workout_count", len(result.Workouts),
				"candidate_count", len(result.CandidateExercises),
				"request_id", request.GetRequestID(ctx),
			)
			return result, nil
		},
	)
}

func runGetWorkoutsTool(ctx context.Context, reader ChatDataReader, input GetWorkoutsToolInput) *GetWorkoutsToolResult {
	if reader == nil {
		return &GetWorkoutsToolResult{Message: "Workout history is not available in this chat."}
	}
	userID, ok := user.Current(ctx)
	if !ok || strings.TrimSpace(userID) == "" {
		return &GetWorkoutsToolResult{Message: "Workout history is not available because this chat has no authenticated user."}
	}

	filter := WorkoutHistoryFilter{
		LastN:        input.LastN,
		WorkoutFocus: strings.TrimSpace(input.WorkoutFocus),
	}
	var notes []string
	if start, note := parseChatToolDate(input.StartDate, false); start != nil {
		filter.StartDate = start
	} else if note != "" {
		notes = append(notes, note)
	}
	if end, note := parseChatToolDate(input.EndDate, true); end != nil {
		filter.EndDate = end
	} else if note != "" {
		notes = append(notes, note)
	}
	if filter.StartDate != nil && filter.EndDate != nil && filter.StartDate.After(*filter.EndDate) {
		filter.StartDate, filter.EndDate = filter.EndDate, filter.StartDate
		notes = append(notes, "The startDate was after the endDate, so I swapped the date filters.")
	}

	if exerciseQuery := strings.TrimSpace(input.ExerciseName); exerciseQuery != "" {
		names, err := reader.ResolveExerciseNames(ctx, userID, exerciseQuery)
		if err != nil {
			return &GetWorkoutsToolResult{Message: appendToolMessage(notes, "I couldn't resolve that exercise name right now.")}
		}
		resolved, candidates, ok := resolveExerciseNameForChat(exerciseQuery, names)
		switch {
		case ok:
			filter.ExerciseName = resolved
		case len(candidates) > 0:
			return &GetWorkoutsToolResult{
				Message:            appendToolMessage(notes, "I found multiple matching exercises. Ask the user which one they mean or retry with an exact exercise name."),
				CandidateExercises: candidates,
			}
		default:
			return &GetWorkoutsToolResult{Message: appendToolMessage(notes, fmt.Sprintf("No exercises matched %q.", exerciseQuery))}
		}
	}

	workouts, err := reader.ListWorkoutsWithSets(ctx, userID, filter)
	if err != nil {
		return &GetWorkoutsToolResult{Message: appendToolMessage(notes, "I couldn't read workout history right now.")}
	}
	if len(workouts) == 0 {
		return &GetWorkoutsToolResult{Message: appendToolMessage(notes, emptyWorkoutMessage(ctx, reader, userID))}
	}
	return &GetWorkoutsToolResult{
		Workouts: workouts,
		Message:  strings.Join(notes, " "),
	}
}

func resolveExerciseNameForChat(query string, matches []string) (string, []string, bool) {
	query = strings.TrimSpace(query)
	if query == "" || len(matches) == 0 {
		return "", nil, false
	}
	for _, match := range matches {
		if strings.EqualFold(strings.TrimSpace(match), query) {
			return match, nil, true
		}
	}
	if len(matches) == 1 {
		return matches[0], nil, true
	}
	return "", append([]string(nil), matches...), false
}

func parseChatToolDate(value string, endOfDay bool) (*time.Time, string) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil, ""
	}

	if parsed, err := time.Parse(time.RFC3339, trimmed); err == nil {
		return &parsed, ""
	}

	dateLayouts := []string{
		"2006-01-02",
		"2006-1-2",
		"January 2, 2006",
		"Jan 2, 2006",
	}
	for _, layout := range dateLayouts {
		parsed, err := time.ParseInLocation(layout, trimmed, time.UTC)
		if err != nil {
			continue
		}
		if endOfDay {
			parsed = parsed.Add(24*time.Hour - time.Nanosecond)
		}
		return &parsed, ""
	}

	return nil, fmt.Sprintf("Ignored invalid date %q.", trimmed)
}

func emptyWorkoutMessage(ctx context.Context, reader ChatDataReader, userID string) string {
	message := "No workouts found for that filter."
	snapshot, err := reader.TrainingSnapshot(ctx, userID)
	if err != nil || snapshot == nil || strings.TrimSpace(snapshot.LastWorkoutDate) == "" {
		return message
	}
	return message + " The user's last workout was " + snapshot.LastWorkoutDate + "."
}

func appendToolMessage(notes []string, message string) string {
	parts := make([]string, 0, len(notes)+1)
	for _, note := range notes {
		if trimmed := strings.TrimSpace(note); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	if trimmed := strings.TrimSpace(message); trimmed != "" {
		parts = append(parts, trimmed)
	}
	return strings.Join(parts, " ")
}
