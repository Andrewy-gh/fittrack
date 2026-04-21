package aichat

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Andrewy-gh/fittrack/server/internal/workout"
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/go-playground/validator/v10"
)

const (
	workoutDraftToolName               = "fittrack.generate_workout_draft"
	workoutDraftSummaryMessage         = "I put together a structured workout draft for you."
	workoutChatFollowUpQuestionCeiling = 3
)

var workoutDraftValidator = validator.New()

type WorkoutGenerationToolInput struct {
	FitnessLevel     string `json:"fitnessLevel" jsonschema:"description=Training experience level. Use beginner, intermediate, or advanced."`
	FitnessGoal      string `json:"fitnessGoal" jsonschema:"description=Primary training goal such as strength, hypertrophy, endurance, power, or general fitness."`
	Equipment        string `json:"equipment" jsonschema:"description=Available equipment such as bodyweight, dumbbells, barbells, machines, cables, or bands."`
	SessionDuration  int    `json:"sessionDuration" jsonschema:"description=Available session length in minutes."`
	WorkoutFocus     string `json:"workoutFocus" jsonschema:"description=Workout focus such as push, pull, legs, upper body, lower body, or full body."`
	SpaceConstraints string `json:"spaceConstraints" jsonschema:"description=Training location or space constraints such as home, gym, hotel, or outdoors."`
	Injuries         string `json:"injuries" jsonschema:"description=Current injuries, pain, or movement limitations. Use none when the user reports no injuries."`
	WorkoutDate      string `json:"workoutDate,omitempty" jsonschema:"description=Optional requested workout date when the user specified one. May be an ISO value or a relative phrase like tomorrow."`
}

func defineWorkoutDraftTool(g *genkit.Genkit, modelName string) ai.Tool {
	return genkit.DefineTool(g, workoutDraftToolName,
		"Creates a FitTrack workout draft. Call this immediately once you know the user's fitness level, goal, equipment, session duration, workout focus, training location, and injury status.",
		func(ctx *ai.ToolContext, input WorkoutGenerationToolInput) (*workout.CreateWorkoutRequest, error) {
			if err := validateWorkoutGenerationToolInput(input); err != nil {
				return nil, err
			}

			return generateWorkoutDraft(ctx, g, modelName, input, time.Now())
		},
	)
}

func validateWorkoutGenerationToolInput(input WorkoutGenerationToolInput) error {
	missing := make([]string, 0, 7)

	if strings.TrimSpace(input.FitnessLevel) == "" {
		missing = append(missing, "fitnessLevel")
	}
	if strings.TrimSpace(input.FitnessGoal) == "" {
		missing = append(missing, "fitnessGoal")
	}
	if strings.TrimSpace(input.Equipment) == "" {
		missing = append(missing, "equipment")
	}
	if input.SessionDuration <= 0 {
		missing = append(missing, "sessionDuration")
	}
	if strings.TrimSpace(input.WorkoutFocus) == "" {
		missing = append(missing, "workoutFocus")
	}
	if strings.TrimSpace(input.SpaceConstraints) == "" {
		missing = append(missing, "spaceConstraints")
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

func generateWorkoutDraft(
	ctx context.Context,
	g *genkit.Genkit,
	modelName string,
	input WorkoutGenerationToolInput,
	now time.Time,
) (*workout.CreateWorkoutRequest, error) {
	output, _, err := genkit.GenerateData[workout.CreateWorkoutRequest](ctx, g,
		ai.WithModelName(modelName),
		ai.WithSystem(buildWorkoutGenerationPrompt(input, now)),
		ai.WithPrompt(buildWorkoutGenerationUserPrompt(input)),
	)
	if err != nil {
		return nil, fmt.Errorf("generate workout draft: %w", err)
	}

	normalizeWorkoutDraft(output)
	if err := validateWorkoutDraft(output); err != nil {
		return nil, fmt.Errorf("validate workout draft: %w", err)
	}

	return output, nil
}

func buildWorkoutGenerationPrompt(input WorkoutGenerationToolInput, now time.Time) string {
	currentTimestamp := now.UTC().Format(time.RFC3339)
	dateInstruction := fmt.Sprintf("If the user did not specify a workout date, use %s for the date field.", currentTimestamp)
	if requested := strings.TrimSpace(input.WorkoutDate); requested != "" {
		dateInstruction = fmt.Sprintf(
			`Resolve the user's requested workout date "%s" relative to %s and return the final value as an RFC3339 timestamp in the "date" field.`,
			requested,
			currentTimestamp,
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
- Do not invent equipment the user does not have.
- Use real, established exercise names only.
- Add weights only when they are reasonably known from the user's context. Otherwise omit weight.
- Keep notes brief and practical. Use notes for rest, effort, or injury reminders when helpful.
- Place compound lifts before accessories when that fits the request.
- %s`, dateInstruction)
}

func buildWorkoutGenerationUserPrompt(input WorkoutGenerationToolInput) string {
	var builder strings.Builder

	builder.WriteString("Build a FitTrack workout draft for this user:\n")
	builder.WriteString(fmt.Sprintf("- Fitness level: %s\n", strings.TrimSpace(input.FitnessLevel)))
	builder.WriteString(fmt.Sprintf("- Goal: %s\n", strings.TrimSpace(input.FitnessGoal)))
	builder.WriteString(fmt.Sprintf("- Equipment: %s\n", strings.TrimSpace(input.Equipment)))
	builder.WriteString(fmt.Sprintf("- Session duration: %d minutes\n", input.SessionDuration))
	builder.WriteString(fmt.Sprintf("- Workout focus: %s\n", strings.TrimSpace(input.WorkoutFocus)))
	builder.WriteString(fmt.Sprintf("- Training location or space: %s\n", strings.TrimSpace(input.SpaceConstraints)))
	builder.WriteString(fmt.Sprintf("- Injuries or limitations: %s\n", strings.TrimSpace(input.Injuries)))
	if requestedDate := strings.TrimSpace(input.WorkoutDate); requestedDate != "" {
		builder.WriteString(fmt.Sprintf("- Requested workout date: %s\n", requestedDate))
	}

	return builder.String()
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

	draft.Date = strings.TrimSpace(draft.Date)
	draft.Notes = trimOptionalString(draft.Notes)
	draft.WorkoutFocus = trimOptionalString(draft.WorkoutFocus)

	for exerciseIndex := range draft.Exercises {
		draft.Exercises[exerciseIndex].Name = strings.TrimSpace(draft.Exercises[exerciseIndex].Name)
		for setIndex := range draft.Exercises[exerciseIndex].Sets {
			draft.Exercises[exerciseIndex].Sets[setIndex].SetType = strings.ToLower(strings.TrimSpace(draft.Exercises[exerciseIndex].Sets[setIndex].SetType))
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

func trimOptionalString(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
