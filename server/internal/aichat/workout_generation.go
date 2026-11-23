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
	workoutDraftToolDescription        = "Creates a FitTrack workout draft. Call this immediately once you know the user's workout focus, session duration, enough equipment or workout context to choose feasible exercises, and injury status. After one follow-up answer, if workout focus, duration, equipment or location context, and usable injury details are present, optional details like fitness level, preferred exercises, or other movements to avoid should not block tool use. Treat stated equipment, location, and space constraints as sufficient and closed by default: use only that context unless the user explicitly mentions more. Do not ask what other equipment they have unless the requested workout is unsafe, contradictory, or not reasonably buildable from those constraints. Do not assume unmentioned accessories or equipment such as a bench, rack, cable, or machine. Equipment is optional for mobility, rehab, prehab, stretching, or warm-up requests because they can default to no-equipment bodyweight work. Use injuries=none only with confirmed user or profile information. If the injury question goes unanswered, ask again and wait before drafting."
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

type workoutDraftLoadChoice struct {
	Count     int      `json:"count,omitempty" jsonschema:"description=Exact number of sets explicitly requested for this scope. Use a separate structure choice with setType working for a working-set count. Omit on positional choices or historical reports."`
	Equipment string   `json:"equipment,omitempty" jsonschema:"description=Equipment confirmed for this set scope. Use an independent structure entry with setType all only when confirmed across the entire exercise."`
	Exercise  string   `json:"exercise" jsonschema:"description=Exact exercise name to use in the draft."`
	Kind      string   `json:"kind" jsonschema:"description=Use requested for a load explicitly requested for this draft. Use reported for performance in an earlier workout. Use blank for a deliberate blank. Use structure for explicitly requested sets, reps, or equipment without a weight instruction."`
	Weight    *float64 `json:"weight,omitempty" jsonschema:"description=Use the user-supplied load in pounds. Use zero for explicitly no added weight. Convert kilograms to pounds. Never infer a number from fitness level or an equipment limit."`
	Reps      int      `json:"reps,omitempty" jsonschema:"description=Reps completed at a reported load. For requested, blank, and structure use the target reps only if specified."`
	SetType   string   `json:"setType,omitempty" jsonschema:"description=Use working by default. Use warmup for explicit warm-up instructions or all for instructions applying to all sets."`
	Last      bool     `json:"last,omitempty" jsonschema:"description=Use for an instruction scoped to the last matching set."`
	Set       int      `json:"set,omitempty" jsonschema:"description=One-based position within the stated set type, or absolute position if no type or all. Omit unless the user requests a specific set position."`
}

type workoutDraftWeightLimit struct {
	Fixed     bool     `json:"fixed,omitempty" jsonschema:"description=True for fixed-weight equipment. Suppress inferred loads and progression."`
	Equipment string   `json:"equipment,omitempty" jsonschema:"description=Equipment class affected by this ceiling such as dumbbell or kettlebell. Include it for a limit that applies to every exercise with that equipment."`
	Exercise  string   `json:"exercise,omitempty" jsonschema:"description=Exact exercise name when the limit is specific to that exercise. May be empty for an equipment-class limit."`
	Weight    *float64 `json:"weight,omitempty" jsonschema:"description=Maximum available load in pounds when explicitly stated. Omit when only a fixed equipment constraint is known."`
}

type WorkoutGenerationToolInput struct {
	InjuryEvidence       string                    `json:"injuryEvidence,omitempty" jsonschema:"description=For current injuries or movement limitations in chat, copy the exact user-authored words reporting the limitation. Do not quote the assistant. Omit when using the saved profile or when the user confirms none."`
	ExerciseWeightLimits []workoutDraftWeightLimit `json:"exerciseWeightLimits,omitempty" jsonschema:"description=Copy explicitly stated limits for each affected equipment class or specific exercise. If the affected equipment cannot be determined set suppressSuggestedWeights."`

	BodyweightOnly           bool                     `json:"bodyweightOnly,omitempty" jsonschema:"description=Set true when the user wants an entirely bodyweight-only workout. Do not invent zero-valued entries for each exercise."`
	SuppressSuggestedWeights bool                     `json:"suppressSuggestedWeights,omitempty" jsonschema:"description=True when the user wants blank weights across the draft or equipment uncertainty makes suggested loads unsuitable. Explicit requested weights still apply."`
	SkipHistoryWeights       bool                     `json:"skipHistoryWeights,omitempty" jsonschema:"description=Set true only to skip FitTrack logged-history lookup. This does not suppress comparable loads reported directly by the user."`
	LoadChoices              []workoutDraftLoadChoice `json:"loadChoices,omitempty" jsonschema:"description=Include explicit set structure, equipment assignments, and user-stated load choices. Include one entry per scoped choice. Exclude weights from workout-history tool calls and any guesses."`
	MaxWeight                *float64                 `json:"maxWeight,omitempty" jsonschema:"description=Only when the user stated a numeric equipment limit that applies across the workout, in pounds. Convert kilograms to pounds. This is a ceiling, never a suggested load."`

	FitnessLevel      string `json:"fitnessLevel,omitempty" jsonschema:"description=Optional training experience level. Use beginner, intermediate, or advanced when known."`
	FitnessGoal       string `json:"fitnessGoal,omitempty" jsonschema:"description=Optional primary training goal such as strength, hypertrophy, endurance, power, or general fitness when known."`
	Equipment         string `json:"equipment,omitempty" jsonschema:"description=Available equipment such as bodyweight, dumbbells, barbells, machines, cables, or bands when known."`
	SessionDuration   int    `json:"sessionDuration" jsonschema:"description=Available session length in minutes."`
	WorkoutFocus      string `json:"workoutFocus" jsonschema:"description=Workout focus such as push, pull, legs, upper body, lower body, or full body."`
	SpaceConstraints  string `json:"spaceConstraints,omitempty" jsonschema:"description=Training location or space constraints such as home, gym, hotel, or outdoors when known."`
	Injuries          string `json:"injuries,omitempty" jsonschema:"description=Current injuries, pain, or movement limitations. Use none when the user reports no injuries."`
	WorkoutDate       string `json:"workoutDate,omitempty" jsonschema:"description=Only if the user requested a date: today, tomorrow, yesterday, YYYY-MM-DD in the user's local day, or an RFC3339 timestamp for an explicitly timed workout."`
	RecentPerformance string `json:"recentPerformance,omitempty" jsonschema:"description=Optional summary of results from FitTrack history tools. Put any weights or sets reported directly by the user into loadChoices instead."`
}

type workoutDraftInputCorrection struct {
	InputCorrection string `json:"inputCorrection"`
}

func defineWorkoutDraftTool(g *genkit.Genkit, modelName string, history draftWeightHistoryReader) ai.Tool {
	return genkit.DefineTool(g, workoutDraftToolName,
		workoutDraftToolDescription,
		func(ctx *ai.ToolContext, input WorkoutGenerationToolInput) (any, error) {
			startedAt := time.Now()
			// Trace marker: shows when the outer chat model entered the workout draft tool.
			logAIChatTraceContext(ctx, "workout_draft_tool_started",
				"model", modelName,
				"session_duration", input.SessionDuration,
				"workout_focus", strings.TrimSpace(input.WorkoutFocus),
				"request_id", request.GetRequestID(ctx),
			)
			now := draftClock(ctx)
			anchor := now
			if source, ok := ctx.Value(draftInjuryContextKey{}).(draftInjuryContext); ok {
				input.Injuries = source.confirmedStatus(input.InjuryEvidence)
				anchor = relativeWorkoutDraftAnchor(input.WorkoutDate, source, now)
			}
			if !hasUsableDraftInjuryStatus(input.Injuries) {
				return workoutDraftInputCorrection{
					InputCorrection: "The injuries field was omitted or does not contain confirmed information. If the conversation or profile establishes injury status, retry the draft tool with that status and the other original inputs. Copy exact user words about any reported limitation into injuryEvidence. Otherwise ask the user to confirm injury status. Never infer none from a missing field.",
				}, nil
			}
			if err := validateWorkoutGenerationToolInput(input); err != nil {
				return nil, err
			}

			schedule, ok := resolveWorkoutDraftSchedule(input.WorkoutDate, now, anchor)
			if !ok {
				return workoutDraftInputCorrection{InputCorrection: "Use today, tomorrow, yesterday, YYYY-MM-DD, or RFC3339 for workoutDate. Ask the user only if the requested date is ambiguous."}, nil
			}
			draft, err := generateWorkoutDraftWithSchedule(ctx, g, modelName, input, schedule, generateWorkoutDraftData)
			if err == nil {
				applyDraftWeights(ctx, history, input, draft, now, schedule.historyThrough)
			}
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

func hasUsableDraftInjuryStatus(value string) bool {
	switch normalizeQualityText(value) {
	case "", "unknown", "not provided", "not specified", "unspecified", "unconfirmed", "not confirmed":
		return false
	default:
		return true
	}
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
	if !hasUsableDraftInjuryStatus(input.Injuries) {
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

// Generated drafts cannot supply catalog links or recommendation snapshots: those
// references come from the exercise catalog and user-specific application state.
// Stored drafts and workout requests retain their existing, broader shape.
type generatedWorkoutDraft struct {
	Date         string                     `json:"date,omitempty"` // Ignored; the server sets the actual date.
	Notes        *string                    `json:"notes,omitempty"`
	WorkoutFocus *string                    `json:"workoutFocus,omitempty"`
	Exercises    []generatedWorkoutExercise `json:"exercises"`
}

type generatedWorkoutExercise struct {
	Name string                `json:"name"`
	Sets []generatedWorkoutSet `json:"sets"`
}

type generatedWorkoutSet struct {
	Reps    int    `json:"reps"`
	SetType string `json:"setType"`
}

func generateWorkoutDraftData(
	ctx context.Context,
	g *genkit.Genkit,
	modelName string,
	systemPrompt string,
	userPrompt string,
) (*workout.CreateWorkoutRequest, error) {
	output, _, err := genkit.GenerateData[generatedWorkoutDraft](ctx, g,
		ai.WithModelName(modelName),
		ai.WithSystem(systemPrompt),
		ai.WithPrompt(userPrompt),
	)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}
	draft := &workout.CreateWorkoutRequest{
		Notes: output.Notes, WorkoutFocus: output.WorkoutFocus,
		Exercises: make([]workout.ExerciseInput, len(output.Exercises)),
	}
	for i, exercise := range output.Exercises {
		sets := make([]workout.SetInput, len(exercise.Sets))
		for j, set := range exercise.Sets {
			sets[j] = workout.SetInput{Reps: set.Reps, SetType: set.SetType}
		}
		draft.Exercises[i] = workout.ExerciseInput{Name: exercise.Name, Sets: sets}
	}
	return draft, nil
}

func generateWorkoutDraftWith(
	ctx context.Context,
	g *genkit.Genkit,
	modelName string,
	input WorkoutGenerationToolInput,
	now time.Time,
	generator workoutDraftGenerator,
) (*workout.CreateWorkoutRequest, error) {
	schedule, ok := resolveWorkoutDraftSchedule(input.WorkoutDate, now, now)
	if !ok {
		return nil, fmt.Errorf("invalid requested workout date")
	}
	return generateWorkoutDraftWithSchedule(ctx, g, modelName, input, schedule, generator)
}

func generateWorkoutDraftWithSchedule(ctx context.Context, g *genkit.Genkit, modelName string, input WorkoutGenerationToolInput, schedule workoutDraftSchedule, generator workoutDraftGenerator) (*workout.CreateWorkoutRequest, error) {
	systemPrompt := buildWorkoutGenerationPrompt()
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

		if output != nil {
			output.Date = schedule.date.Format(time.RFC3339)
		}
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

func buildWorkoutGenerationPrompt() string {
	return `You are generating a FitTrack workout draft.

Return structured data only, with the following fields:
{
  "notes": "optional string",
  "workoutFocus": "optional string",
  "exercises": [
    {
      "name": "exercise name",
      "sets": [
        {
          "reps": 8,
          "setType": "warmup" | "working"
        }
      ]
    }
  ]
}

Rules:
- Include every required field. FitTrack attaches the actual workout date independently of the model.
- The exercise list must contain at least one real exercise, and every exercise must contain at least one set.
- Use only "warmup" or "working" for setType.
- Match the workout focus, available equipment, session length, training location, and injury constraints.
- Scale the draft to the requested session duration by estimating setup and transitions, set execution time, rest between sets, and warm-up or ramp-up needs when appropriate.
- Use a reasonable share of the requested time. Do not satisfy a normal 40+ minute strength or hypertrophy request with a very small workout unless the user asked for minimal, beginner, rehab, warm-up, or low-volume work.
- Adjust density by goal: strength can use fewer exercises with longer rests and enough sets; hypertrophy should use moderate rests and enough total working sets; endurance or circuit work should use shorter rests and higher density; rehab, mobility, and beginner sessions can use lower volume when appropriate.
- Do not invent equipment the user does not have.
- Use real, established exercise names only.
- When recent performance is supplied, use it to choose suitable exercises and reps.
- Do not assign any weights. FitTrack applies user-specified loads and history after selecting the exercises and sets.
- Keep notes brief and practical. Use notes for rest, effort, or injury reminders when helpful. Do not prescribe numeric weights in notes.
- Place compound lifts before accessories when that fits the request.
`
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
	if recentPerformance := strings.TrimSpace(input.RecentPerformance); recentPerformance != "" {
		builder.WriteString(fmt.Sprintf("- Recent performance to respect: %s\n", recentPerformance))
	}

	for _, choice := range input.LoadChoices {
		if strings.TrimSpace(choice.Exercise) == "" {
			continue
		}
		builder.WriteString(fmt.Sprintf("- Include %q using that exact name if feasible and compatible with the request.\n", choice.Exercise))
		if choice.Kind == "structure" && choice.Count > 0 && choice.Set == 0 && !choice.Last {
			builder.WriteString(fmt.Sprintf("  Include exactly %d sets within %s for %q.\n", choice.Count, draftLoadScope(choice), choice.Exercise))
		}
		if choice.Kind == "requested" || choice.Kind == "structure" {
			if choice.Reps > 0 {
				builder.WriteString(fmt.Sprintf("  Prescribe %d reps for %s of %q.\n", choice.Reps, draftLoadScope(choice), choice.Exercise))
			} else {
				builder.WriteString(fmt.Sprintf("  Include %s for %q. Choose appropriate reps.\n", draftLoadScope(choice), choice.Exercise))
			}
		} else if choice.Kind == "blank" && choice.hasSetScope() {
			builder.WriteString(fmt.Sprintf("  Include %s for %q with weight left blank. Preserve any rep target specified for those sets.\n", draftLoadScope(choice), choice.Exercise))
			if choice.Reps > 0 {
				builder.WriteString(fmt.Sprintf("  Prescribe %d reps for those sets.\n", choice.Reps))
			}
		} else if choice.Kind == "reported" && choice.Reps > 0 {
			builder.WriteString(fmt.Sprintf("  Use %d reps for %s if including those sets, unless the user requests a different target.\n", choice.Reps, draftLoadScope(choice)))
		}
	}
	for _, limit := range input.ExerciseWeightLimits {
		if limit.Exercise != "" {
			builder.WriteString(fmt.Sprintf("- Use the exact name %q if including that equipment-limited exercise.\n", limit.Exercise))
		}
		if limit.Equipment != "" {
			builder.WriteString(fmt.Sprintf("- Include equipment type %q in the name of each exercise that uses it.\n", limit.Equipment))
		}
	}
	return builder.String()
}

func draftLoadScope(choice workoutDraftLoadChoice) string {
	setType := draftLoadSetType(choice)
	if setType == "" || setType == "all" {
		setType = "workout"
	}
	if choice.Set > 0 {
		return fmt.Sprintf("set %d within the ordered %s sets", choice.Set, setType)
	}
	if choice.Last {
		return "the last " + setType + " set"
	}
	return "the requested " + setType + " sets"
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

			var result struct {
				workoutDraftInputCorrection
				workout.CreateWorkoutRequest
			}
			if err := json.Unmarshal(payload, &result); err != nil {
				return nil, fmt.Errorf("decode workout draft tool output: %w", err)
			}
			if result.InputCorrection != "" {
				continue
			}
			draft := result.CreateWorkoutRequest
			normalizeWorkoutDraft(&draft)
			if err := validateWorkoutDraft(&draft); err != nil {
				return nil, err
			}

			return &draft, nil
		}
	}

	return nil, nil
}

func extractToolCallsFromHistory(history []*ai.Message) []string {
	var calls []string
	for _, message := range history {
		if message == nil {
			continue
		}
		for _, part := range message.Content {
			if part == nil || !part.IsToolRequest() || part.ToolRequest == nil {
				continue
			}
			if name := strings.TrimSpace(part.ToolRequest.Name); name != "" {
				calls = append(calls, name)
			}
		}
	}
	return calls
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
