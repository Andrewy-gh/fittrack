package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-playground/validator/v10"
)

const (
	workoutDraftToolName               = "generate_workout_draft"
	workoutDraftToolDescription        = "Creates a FitTrack workout draft. Call this immediately once you know the user's workout focus, session duration, enough equipment or workout context to choose feasible exercises, and injury status. Equipment is optional for mobility, rehab, prehab, stretching, or warm-up requests because they can default to no-equipment bodyweight work. Use injuries=none only when the user explicitly reports no injuries or continues without answering after one injury-status question. Fitness level improves weight assumptions but is not required."
	workoutDraftSummaryMessage         = "I put together a structured workout draft for you."
	workoutChatFollowUpQuestionCeiling = 3
)

var workoutDraftValidator = validator.New()

type workoutDraftGenerator func(
	ctx context.Context,
	g *genkit.Genkit,
	modelName string,
	systemPrompt string,
	userPrompt string,
) (*workout.CreateWorkoutRequest, error)

type WorkoutGenerationToolInput struct {
	FitnessLevel     string `json:"fitnessLevel,omitempty" jsonschema:"description=Optional training experience level. Use beginner, intermediate, or advanced when known."`
	FitnessGoal      string `json:"fitnessGoal,omitempty" jsonschema:"description=Optional primary training goal such as strength, hypertrophy, endurance, power, or general fitness when known."`
	Equipment        string `json:"equipment,omitempty" jsonschema:"description=Available equipment such as bodyweight, dumbbells, barbells, machines, cables, or bands when known."`
	SessionDuration  int    `json:"sessionDuration" jsonschema:"description=Available session length in minutes."`
	WorkoutFocus     string `json:"workoutFocus" jsonschema:"description=Workout focus such as push, pull, legs, upper body, lower body, or full body."`
	SpaceConstraints string `json:"spaceConstraints,omitempty" jsonschema:"description=Training location or space constraints such as home, gym, hotel, or outdoors when known."`
	Injuries         string `json:"injuries" jsonschema:"description=Current injuries, pain, or movement limitations. Use none when the user reports no injuries."`
	WorkoutDate      string `json:"workoutDate,omitempty" jsonschema:"description=Optional requested workout date when the user specified one. May be an ISO value or a relative phrase like tomorrow."`
}

func defineWorkoutDraftTool(g *genkit.Genkit, modelName string) ai.Tool {
	return genkit.DefineTool(g, workoutDraftToolName,
		workoutDraftToolDescription,
		func(ctx *ai.ToolContext, input WorkoutGenerationToolInput) (*workout.CreateWorkoutRequest, error) {
			startedAt := time.Now()
			// Trace marker: shows when the outer chat model entered the workout draft tool.
			logAIChatTraceContext(ctx, "workout_draft_tool_started",
				"model", modelName,
				"session_duration", input.SessionDuration,
				"workout_focus", strings.TrimSpace(input.WorkoutFocus),
				"request_id", request.GetRequestID(ctx),
			)
			if err := validateWorkoutGenerationToolInput(input); err != nil {
				return nil, err
			}

			draft, err := generateWorkoutDraft(ctx, g, modelName, input, time.Now())
			logAIChatTraceContext(ctx, "workout_draft_tool_finished",
				"elapsed_ms", time.Since(startedAt).Milliseconds(),
				"model", modelName,
				"session_duration", input.SessionDuration,
				"workout_focus", strings.TrimSpace(input.WorkoutFocus),
				"exercise_count", workoutDraftExerciseCount(draft),
				"error", traceError(err),
				"request_id", request.GetRequestID(ctx),
			)
			return draft, err
		},
	)
}

func validateWorkoutGenerationToolInput(input WorkoutGenerationToolInput) error {
	missing := make([]string, 0, 4)

	if input.SessionDuration <= 0 {
		missing = append(missing, "sessionDuration")
	}
	if strings.TrimSpace(input.WorkoutFocus) == "" {
		missing = append(missing, "workoutFocus")
	}
	if !hasWorkoutContext(input) {
		missing = append(missing, "equipment or spaceConstraints")
	}
	if strings.TrimSpace(input.Injuries) == "" {
		missing = append(missing, "injuries")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required workout generation fields: %s", strings.Join(missing, ", "))
	}
	if input.SessionDuration > 300 {
		return fmt.Errorf("sessionDuration must be between 1 and 300 minutes")
	}

	return nil
}

func hasWorkoutContext(input WorkoutGenerationToolInput) bool {
	return strings.TrimSpace(input.Equipment) != "" ||
		strings.TrimSpace(input.SpaceConstraints) != "" ||
		isEquipmentOptionalWorkout(input)
}

func isEquipmentOptionalWorkout(input WorkoutGenerationToolInput) bool {
	text := normalizeQualityText(input.FitnessGoal + " " + input.WorkoutFocus)
	return hasAnyQualityTerm(text, "mobility", "rehab", "prehab", "stretching", "warm-up", "warmup")
}

func generateWorkoutDraft(
	ctx context.Context,
	g *genkit.Genkit,
	modelName string,
	input WorkoutGenerationToolInput,
	now time.Time,
) (*workout.CreateWorkoutRequest, error) {
	return generateWorkoutDraftWith(ctx, g, modelName, input, now, generateWorkoutDraftData)
}

func generateWorkoutDraftData(
	ctx context.Context,
	g *genkit.Genkit,
	modelName string,
	systemPrompt string,
	userPrompt string,
) (*workout.CreateWorkoutRequest, error) {
	output, _, err := genkit.GenerateData[workout.CreateWorkoutRequest](ctx, g,
		ai.WithModelName(modelName),
		ai.WithSystem(systemPrompt),
		ai.WithPrompt(userPrompt),
	)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func generateWorkoutDraftWith(
	ctx context.Context,
	g *genkit.Genkit,
	modelName string,
	input WorkoutGenerationToolInput,
	now time.Time,
	generator workoutDraftGenerator,
) (*workout.CreateWorkoutRequest, error) {
	systemPrompt := buildWorkoutGenerationPrompt(input, now)
	userPrompt := buildWorkoutGenerationUserPrompt(input)

	for attempt := 1; attempt <= 2; attempt++ {
		attemptStartedAt := time.Now()
		// Trace marker: each attempt is a structured model call; a second attempt means quality repair ran.
		logAIChatTraceContext(ctx, "workout_draft_generation_attempt_started",
			"attempt", attempt,
			"model", modelName,
			"session_duration", input.SessionDuration,
			"workout_focus", strings.TrimSpace(input.WorkoutFocus),
			"request_id", request.GetRequestID(ctx),
		)
		output, err := generator(ctx, g, modelName, systemPrompt, userPrompt)
		if err != nil {
			logAIChatTraceContext(ctx, "workout_draft_generation_attempt_finished",
				"attempt", attempt,
				"elapsed_ms", time.Since(attemptStartedAt).Milliseconds(),
				"model", modelName,
				"error", traceError(err),
				"request_id", request.GetRequestID(ctx),
			)
			return nil, fmt.Errorf("generate workout draft: %w", err)
		}
		logAIChatTraceContext(ctx, "workout_draft_generation_attempt_finished",
			"attempt", attempt,
			"elapsed_ms", time.Since(attemptStartedAt).Milliseconds(),
			"model", modelName,
			"exercise_count", workoutDraftExerciseCount(output),
			"request_id", request.GetRequestID(ctx),
		)

		normalizeWorkoutDraft(output)
		if err := validateWorkoutDraft(output); err != nil {
			return nil, fmt.Errorf("validate workout draft: %w", err)
		}

		qualityErr := validateWorkoutDraftQuality(input, output)
		if qualityErr == nil {
			return output, nil
		}
		if attempt == 2 {
			return nil, fmt.Errorf("validate workout draft quality after repair retry: %w", qualityErr)
		}

		// Trace marker: explains why the tool is about to spend time on a second model call.
		logAIChatTraceContext(ctx, "workout_draft_quality_retry",
			"attempt", attempt,
			"model", modelName,
			"reason", qualityErr.Error(),
			"request_id", request.GetRequestID(ctx),
		)
		userPrompt = buildWorkoutGenerationRepairPrompt(input, qualityErr)
	}

	return nil, fmt.Errorf("generate workout draft: exhausted quality repair attempts")
}

func workoutDraftExerciseCount(draft *workout.CreateWorkoutRequest) int {
	if draft == nil {
		return 0
	}
	return len(draft.Exercises)
}

func traceError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func buildWorkoutGenerationPrompt(input WorkoutGenerationToolInput, now time.Time) string {
	currentTimestamp := now.Format(time.RFC3339)
	dateInstruction := fmt.Sprintf("If the user did not specify a workout date, use %s for the date field.", currentTimestamp)
	if requested := strings.TrimSpace(input.WorkoutDate); requested != "" {
		dateInstruction = fmt.Sprintf(
			`Use the user's requested workout date "%s" for the "date" field. If the request is relative, interpret it from the user's local day rather than UTC, and do not shift it because of a UTC date boundary. Return the final value as an RFC3339 timestamp in the "date" field.`,
			requested,
		)
	}

	return fmt.Sprintf(`You are generating a FitTrack workout draft.

Return structured data only. The output must match FitTrack's current workout create request exactly:
{
  "date": "RFC3339 timestamp",
  "notes": "optional string",
  "workoutFocus": "optional string",
  "exercises": [
    {
      "name": "exercise name",
      "sets": [
        {
          "weight": 135,
          "reps": 8,
          "setType": "warmup" | "working"
        }
      ]
    }
  ]
}

Rules:
- Include every required field. "date" is always required and must be RFC3339.
- The exercise list must contain at least one real exercise, and every exercise must contain at least one set.
- Use only "warmup" or "working" for setType.
- Match the workout focus, available equipment, session length, training location, and injury constraints.
- Scale the draft to the requested session duration by estimating setup and transitions, set execution time, rest between sets, and warm-up or ramp-up needs when appropriate.
- Use a reasonable share of the requested time. Do not satisfy a normal 40+ minute strength or hypertrophy request with a very small workout unless the user asked for minimal, beginner, rehab, warm-up, or low-volume work.
- Adjust density by goal: strength can use fewer exercises with longer rests and enough sets; hypertrophy should use moderate rests and enough total working sets; endurance or circuit work should use shorter rests and higher density; rehab, mobility, and beginner sessions can use lower volume when appropriate.
- Do not invent equipment the user does not have.
- Use real, established exercise names only.
- Add weights only when they are reasonably known from the user's context. If fitness level is unknown, prefer omitting weights instead of guessing aggressively.
- Keep notes brief and practical. Use notes for rest, effort, or injury reminders when helpful.
- Place compound lifts before accessories when that fits the request.
- %s`, dateInstruction)
}

func buildWorkoutGenerationUserPrompt(input WorkoutGenerationToolInput) string {
	var builder strings.Builder

	builder.WriteString("Build a FitTrack workout draft for this user:\n")
	builder.WriteString(fmt.Sprintf("- Fitness level: %s\n", workoutPromptValue(input.FitnessLevel)))
	builder.WriteString(fmt.Sprintf("- Goal: %s\n", workoutPromptValue(input.FitnessGoal)))
	builder.WriteString(fmt.Sprintf("- Equipment: %s\n", workoutPromptValue(input.Equipment)))
	builder.WriteString(fmt.Sprintf("- Session duration: %d minutes\n", input.SessionDuration))
	builder.WriteString(fmt.Sprintf("- Workout focus: %s\n", workoutPromptValue(input.WorkoutFocus)))
	builder.WriteString(fmt.Sprintf("- Training location or space: %s\n", workoutPromptValue(input.SpaceConstraints)))
	builder.WriteString(fmt.Sprintf("- Injuries or limitations: %s\n", workoutPromptValue(input.Injuries)))
	if requestedDate := strings.TrimSpace(input.WorkoutDate); requestedDate != "" {
		builder.WriteString(fmt.Sprintf("- Requested workout date: %s\n", requestedDate))
	}

	return builder.String()
}

func buildWorkoutGenerationRepairPrompt(input WorkoutGenerationToolInput, qualityErr error) string {
	var builder strings.Builder

	builder.WriteString("Regenerate the FitTrack workout draft for the same user.\n")
	builder.WriteString("The previous draft failed deterministic quality validation. Fix these issues:\n")
	for _, issue := range strings.Split(qualityErr.Error(), "; ") {
		issue = strings.TrimSpace(issue)
		if issue == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("- %s\n", issue))
	}
	builder.WriteString("\nReturn a complete replacement draft that still follows the original request:\n")
	builder.WriteString(buildWorkoutGenerationUserPrompt(input))

	return builder.String()
}

func workoutPromptValue(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "unknown"
	}

	return trimmed
}

func validateWorkoutDraft(draft *workout.CreateWorkoutRequest) error {
	if draft == nil {
		return fmt.Errorf("workout draft is required")
	}

	return workoutDraftValidator.Struct(*draft)
}

func normalizeWorkoutDraft(draft *workout.CreateWorkoutRequest) {
	if draft == nil {
		return
	}

	draft.Date = cleanWorkoutDraftText(draft.Date)
	draft.Notes = cleanOptionalWorkoutDraftText(draft.Notes)
	draft.WorkoutFocus = cleanOptionalWorkoutDraftText(draft.WorkoutFocus)

	for exerciseIndex := range draft.Exercises {
		draft.Exercises[exerciseIndex].Name = cleanWorkoutDraftText(draft.Exercises[exerciseIndex].Name)
		for setIndex := range draft.Exercises[exerciseIndex].Sets {
			setType := cleanWorkoutDraftText(draft.Exercises[exerciseIndex].Sets[setIndex].SetType)
			draft.Exercises[exerciseIndex].Sets[setIndex].SetType = strings.ToLower(setType)
		}
	}
}

func extractWorkoutDraftFromHistory(history []*ai.Message) (*workout.CreateWorkoutRequest, error) {
	for messageIndex := len(history) - 1; messageIndex >= 0; messageIndex-- {
		message := history[messageIndex]
		if message == nil || message.Role != ai.RoleTool {
			continue
		}

		for _, part := range message.Content {
			if part == nil || !part.IsToolResponse() || part.ToolResponse == nil {
				continue
			}
			if part.ToolResponse.Name != workoutDraftToolName {
				continue
			}

			payload, err := json.Marshal(part.ToolResponse.Output)
			if err != nil {
				return nil, fmt.Errorf("marshal workout draft tool output: %w", err)
			}

			var draft workout.CreateWorkoutRequest
			if err := json.Unmarshal(payload, &draft); err != nil {
				return nil, fmt.Errorf("decode workout draft tool output: %w", err)
			}

			normalizeWorkoutDraft(&draft)
			if err := validateWorkoutDraft(&draft); err != nil {
				return nil, err
			}

			return &draft, nil
		}
	}

	return nil, nil
}

func finalizeAssistantText(text string, draft *workout.CreateWorkoutRequest) string {
	text = strings.TrimSpace(text)
	if draft == nil {
		return text
	}
	if text == "" || looksLikeExerciseDump(text, draft) {
		return workoutDraftSummaryMessage
	}

	return text
}

func looksLikeExerciseDump(text string, draft *workout.CreateWorkoutRequest) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return false
	}
	if strings.Count(normalized, "\n- ") >= 2 || strings.Count(normalized, "\n1.") >= 1 {
		return true
	}

	for _, exercise := range draft.Exercises {
		name := strings.ToLower(strings.TrimSpace(exercise.Name))
		if name == "" {
			continue
		}
		if strings.Contains(normalized, name) {
			return true
		}
	}

	return false
}

func cleanOptionalWorkoutDraftText(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := cleanWorkoutDraftText(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}

func cleanWorkoutDraftText(value string) string {
	return strings.TrimSpace(strings.ReplaceAll(value, "\x00", ""))
}
