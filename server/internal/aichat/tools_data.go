package aichat

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

const (
	getWorkoutsToolName             = "get_workouts"
	getWorkoutsToolDescription      = "Reads the authenticated user's logged FitTrack workouts. Use this for questions about their personal workout history, recent sessions, exercises performed, or logged training data. Do not use it for general fitness knowledge or to create a new workout draft."
	getExerciseStatsToolName        = "get_exercise_stats"
	getExerciseStatsToolDescription = "Reads compact stats for one of the authenticated user's exercises: all-time best estimated 1RM, recent trend points, last-session sets, and session count. Use this only for all-time bests or long-range single-exercise trends, or before drafting when the user explicitly wants the draft based on past performance."
	updateTrainingProfileToolName   = "update_training_profile"
	updateTrainingProfileToolDesc   = "Updates durable training profile facts for the authenticated user. Use only when the user states a lasting preference, limitation, usual training setup, goal, or asks to forget/clear profile facts. Do not use for one-off details about today's workout."
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

type GetExerciseStatsToolInput struct {
	ExerciseName string `json:"exerciseName" jsonschema:"description=Exercise name to inspect. Partial name is okay, e.g. bench or squat."`
	Window       string `json:"window,omitempty" jsonschema:"description=Stats window: 3m, 1y, or all. Default 3m."`
}

type GetExerciseStatsToolResult struct {
	Stats              *ExerciseStatsView `json:"stats,omitempty"`
	Message            string             `json:"message,omitempty"`
	CandidateExercises []string           `json:"candidateExercises,omitempty"`
}

type UpdateTrainingProfileToolInput struct {
	PrimaryGoal                     *string   `json:"primaryGoal,omitempty" jsonschema:"description=Durable primary goal. Allowed: strength, hypertrophy, endurance, general_fitness, weight_loss, mobility. Empty string clears it."`
	ExperienceLevel                 *string   `json:"experienceLevel,omitempty" jsonschema:"description=Durable experience level. Allowed: beginner, intermediate, advanced. Empty string clears it."`
	PreferredSessionDurationMinutes *int32    `json:"preferredSessionDurationMinutes,omitempty" jsonschema:"description=Durable preferred session duration in minutes, 10 to 240. Zero clears it."`
	UsualTrainingLocation           *string   `json:"usualTrainingLocation,omitempty" jsonschema:"description=Durable usual training location. Allowed: gym, home, outdoor, travel. Empty string clears it."`
	AvailableEquipment              *[]string `json:"availableEquipment,omitempty" jsonschema:"description=Complete durable list of available equipment. Send the full list; empty list clears it."`
	AvoidedExercises                *[]string `json:"avoidedExercises,omitempty" jsonschema:"description=Complete durable list of exercises the user avoids. Send the full list; empty list clears it."`
	MovementLimitations             *[]string `json:"movementLimitations,omitempty" jsonschema:"description=Complete durable list of injuries, movement limits, or none stated. Send the full list; empty list clears it."`
}

type UpdateTrainingProfileToolResult struct {
	Profile       *TrainingProfile `json:"profile,omitempty"`
	UpdatedFields []string         `json:"updatedFields,omitempty"`
	Message       string           `json:"message,omitempty"`
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

func defineGetExerciseStatsTool(g *genkit.Genkit, reader ChatDataReader) ai.Tool {
	return genkit.DefineTool(g, getExerciseStatsToolName,
		getExerciseStatsToolDescription,
		func(ctx *ai.ToolContext, input GetExerciseStatsToolInput) (*GetExerciseStatsToolResult, error) {
			startedAt := time.Now()
			logAIChatTraceContext(ctx, "get_exercise_stats_tool_started",
				"exercise_name", strings.TrimSpace(input.ExerciseName),
				"window", strings.TrimSpace(input.Window),
				"request_id", request.GetRequestID(ctx),
			)
			result := runGetExerciseStatsTool(ctx, reader, input)
			statCount := 0
			if result.Stats != nil {
				statCount = len(result.Stats.Trend)
			}
			logAIChatTraceContext(ctx, "get_exercise_stats_tool_finished",
				"elapsed_ms", time.Since(startedAt).Milliseconds(),
				"trend_count", statCount,
				"candidate_count", len(result.CandidateExercises),
				"request_id", request.GetRequestID(ctx),
			)
			return result, nil
		},
	)
}

func defineUpdateTrainingProfileTool(g *genkit.Genkit, reader ChatDataReader) ai.Tool {
	return genkit.DefineTool(g, updateTrainingProfileToolName,
		updateTrainingProfileToolDesc,
		func(ctx *ai.ToolContext, input UpdateTrainingProfileToolInput) (*UpdateTrainingProfileToolResult, error) {
			startedAt := time.Now()
			result := runUpdateTrainingProfileTool(ctx, reader, input)
			logAIChatTraceContext(ctx, "update_training_profile_tool_finished",
				"elapsed_ms", time.Since(startedAt).Milliseconds(),
				"updated_fields", strings.Join(result.UpdatedFields, ","),
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

func runUpdateTrainingProfileTool(ctx context.Context, reader ChatDataReader, input UpdateTrainingProfileToolInput) *UpdateTrainingProfileToolResult {
	if reader == nil {
		return &UpdateTrainingProfileToolResult{Message: "Training profile updates are not available in this chat."}
	}
	userID, ok := user.Current(ctx)
	if !ok || strings.TrimSpace(userID) == "" {
		return &UpdateTrainingProfileToolResult{Message: "Training profile updates are not available because this chat has no authenticated user."}
	}

	update, fields, notes := buildTrainingProfileUpdate(ctx, input)
	if len(fields) == 0 {
		return &UpdateTrainingProfileToolResult{Message: appendToolMessage(notes, "No supported training profile fields were provided.")}
	}

	profile, err := reader.UpdateTrainingProfile(ctx, userID, update)
	if err != nil {
		return &UpdateTrainingProfileToolResult{Message: appendToolMessage(notes, "I couldn't update the training profile right now.")}
	}
	sort.Strings(fields)
	return &UpdateTrainingProfileToolResult{
		Profile:       profile,
		UpdatedFields: fields,
		Message:       appendToolMessage(notes, "Training profile updated."),
	}
}

func buildTrainingProfileUpdate(ctx context.Context, input UpdateTrainingProfileToolInput) (TrainingProfileUpdate, []string, []string) {
	var update TrainingProfileUpdate
	var fields []string
	var notes []string

	if input.PrimaryGoal != nil {
		if goal, ok := normalizeProfileEnum(*input.PrimaryGoal, profileGoalAliases); ok {
			update.PrimaryGoal = &goal
			fields = append(fields, "primary_goal")
		} else {
			notes = append(notes, fmt.Sprintf("Ignored unsupported primary goal %q.", strings.TrimSpace(*input.PrimaryGoal)))
		}
	}
	if input.ExperienceLevel != nil {
		if level, ok := normalizeProfileEnum(*input.ExperienceLevel, profileExperienceAliases); ok {
			update.ExperienceLevel = &level
			fields = append(fields, "experience_level")
		} else {
			notes = append(notes, fmt.Sprintf("Ignored unsupported experience level %q.", strings.TrimSpace(*input.ExperienceLevel)))
		}
	}
	if input.PreferredSessionDurationMinutes != nil {
		duration := *input.PreferredSessionDurationMinutes
		switch {
		case duration <= 0:
			duration = 0
		case duration < 10:
			duration = 10
			notes = append(notes, "Preferred session duration was raised to the 10 minute minimum.")
		case duration > 240:
			duration = 240
			notes = append(notes, "Preferred session duration was capped at 240 minutes.")
		}
		update.PreferredSessionDurationMinutes = &duration
		fields = append(fields, "preferred_session_duration_minutes")
	}
	if input.UsualTrainingLocation != nil {
		if location, ok := normalizeProfileEnum(*input.UsualTrainingLocation, profileLocationAliases); ok {
			update.UsualTrainingLocation = &location
			fields = append(fields, "usual_training_location")
		} else {
			notes = append(notes, fmt.Sprintf("Ignored unsupported training location %q.", strings.TrimSpace(*input.UsualTrainingLocation)))
		}
	}
	if input.AvailableEquipment != nil {
		values := cleanProfileStringList(*input.AvailableEquipment)
		update.AvailableEquipment = &values
		fields = append(fields, "available_equipment")
	}
	if input.AvoidedExercises != nil {
		values := cleanProfileStringList(*input.AvoidedExercises)
		update.AvoidedExercises = &values
		fields = append(fields, "avoided_exercises")
	}
	if input.MovementLimitations != nil {
		values := cleanProfileStringList(*input.MovementLimitations)
		update.MovementLimitations = &values
		fields = append(fields, "movement_limitations")
	}

	if source, ok := trainingProfileSourceFromContext(ctx); ok {
		update.SourceConversationID = &source.ConversationID
		update.SourceMessageID = &source.MessageID
	}
	return update, fields, notes
}

func runGetExerciseStatsTool(ctx context.Context, reader ChatDataReader, input GetExerciseStatsToolInput) *GetExerciseStatsToolResult {
	if reader == nil {
		return &GetExerciseStatsToolResult{Message: "Exercise stats are not available in this chat."}
	}
	userID, ok := user.Current(ctx)
	if !ok || strings.TrimSpace(userID) == "" {
		return &GetExerciseStatsToolResult{Message: "Exercise stats are not available because this chat has no authenticated user."}
	}

	exerciseQuery := strings.TrimSpace(input.ExerciseName)
	if exerciseQuery == "" {
		return &GetExerciseStatsToolResult{Message: "Ask the user which exercise they want stats for."}
	}

	names, err := reader.ResolveExerciseNames(ctx, userID, exerciseQuery)
	if err != nil {
		return &GetExerciseStatsToolResult{Message: "I couldn't resolve that exercise name right now."}
	}
	resolved, candidates, ok := resolveExerciseNameForChat(exerciseQuery, names)
	switch {
	case ok:
	case len(candidates) > 0:
		return &GetExerciseStatsToolResult{
			Message:            "I found multiple matching exercises. Ask the user which one they mean or retry with an exact exercise name.",
			CandidateExercises: candidates,
		}
	default:
		return &GetExerciseStatsToolResult{Message: fmt.Sprintf("No exercises matched %q.", exerciseQuery)}
	}

	stats, err := reader.ExerciseStats(ctx, userID, resolved, input.Window)
	if err != nil {
		return &GetExerciseStatsToolResult{Message: "I couldn't read exercise stats right now."}
	}
	if stats == nil {
		return &GetExerciseStatsToolResult{Message: "No exercise stats were found."}
	}
	return &GetExerciseStatsToolResult{
		Stats:   stats,
		Message: strings.TrimSpace(stats.Message),
	}
}

var profileGoalAliases = map[string]string{
	"strength":        "strength",
	"hypertrophy":     "hypertrophy",
	"muscle growth":   "hypertrophy",
	"muscle_gain":     "hypertrophy",
	"muscle gain":     "hypertrophy",
	"endurance":       "endurance",
	"general":         "general_fitness",
	"general fitness": "general_fitness",
	"general_fitness": "general_fitness",
	"weight loss":     "weight_loss",
	"weight_loss":     "weight_loss",
	"fat loss":        "weight_loss",
	"fat_loss":        "weight_loss",
	"mobility":        "mobility",
}

var profileExperienceAliases = map[string]string{
	"beginner":     "beginner",
	"intermediate": "intermediate",
	"advanced":     "advanced",
}

var profileLocationAliases = map[string]string{
	"gym":            "gym",
	"commercial gym": "gym",
	"full gym":       "gym",
	"home":           "home",
	"home gym":       "home",
	"outdoor":        "outdoor",
	"outside":        "outdoor",
	"travel":         "travel",
	"hotel":          "travel",
	"hotel gym":      "travel",
}

func normalizeProfileEnum(value string, aliases map[string]string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", true
	}
	normalized := strings.ToLower(strings.ReplaceAll(trimmed, "-", " "))
	normalized = strings.Join(strings.Fields(normalized), " ")
	if mapped, ok := aliases[normalized]; ok {
		return mapped, true
	}
	underscore := strings.ReplaceAll(normalized, " ", "_")
	if mapped, ok := aliases[underscore]; ok {
		return mapped, true
	}
	return "", false
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
