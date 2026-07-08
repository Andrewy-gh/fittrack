package aichat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"

	"github.com/Andrewy-gh/fittrack/server/internal/request"
	"github.com/Andrewy-gh/fittrack/server/internal/user"
)

const (
	geminiAPIKeyEnvVar = "GEMINI_API_KEY"
	googleAPIKeyEnvVar = "GOOGLE_API_KEY"
	chatStreamTimeout  = chatGenerationTimeout
	chatMaxTurns       = 6
	// chatEmptyResponseRetryLimit caps regeneration attempts when the model
	// returns an empty candidate with no text and no tool calls.
	chatEmptyResponseRetryLimit = 1
)

var genkitInit = func(ctx context.Context, opts ...genkit.GenkitOption) *genkit.Genkit {
	return genkit.Init(ctx, opts...)
}

type GenkitRuntime struct {
	available            bool
	modelName            string
	g                    *genkit.Genkit
	workoutDraftTool     ai.Tool
	getWorkoutsTool      ai.Tool
	getExerciseStatsTool ai.Tool
	updateProfileTool    ai.Tool
	dataReader           ChatDataReader
}

func NewGenkitRuntime(ctx context.Context, reader ChatDataReader) *GenkitRuntime {
	modelName := resolveModelName()
	runtime := &GenkitRuntime{
		modelName:  modelName,
		dataReader: reader,
	}

	if configuredAPIKeyEnvVar() == "" {
		return runtime
	}

	g, workoutDraftTool, getWorkoutsTool, getExerciseStatsTool, updateProfileTool, ok := activateGenkitRuntime(ctx, modelName, reader)
	if !ok {
		return runtime
	}

	runtime.g = g
	runtime.workoutDraftTool = workoutDraftTool
	runtime.getWorkoutsTool = getWorkoutsTool
	runtime.getExerciseStatsTool = getExerciseStatsTool
	runtime.updateProfileTool = updateProfileTool
	runtime.available = true

	return runtime
}

func activateGenkitRuntime(ctx context.Context, modelName string, reader ChatDataReader) (_ *genkit.Genkit, _ ai.Tool, _ ai.Tool, _ ai.Tool, _ ai.Tool, ok bool) {
	defer func() {
		if recovered := recover(); recovered != nil {
			slog.Warn("ai chat runtime initialization skipped after genkit panic",
				"configured_api_key_env", googleAPIKeyEnvVar,
				"panic", recovered,
			)
			ok = false
		}
	}()

	g := genkitInit(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel(modelName),
	)

	workoutDraftTool := defineWorkoutDraftTool(g, modelName)
	var getWorkoutsTool ai.Tool
	var getExerciseStatsTool ai.Tool
	var updateProfileTool ai.Tool
	if reader != nil {
		getWorkoutsTool = defineGetWorkoutsTool(g, reader)
		getExerciseStatsTool = defineGetExerciseStatsTool(g, reader)
		updateProfileTool = defineUpdateTrainingProfileTool(g, reader)
	}

	return g, workoutDraftTool, getWorkoutsTool, getExerciseStatsTool, updateProfileTool, true
}

func (r *GenkitRuntime) ModelName() string {
	return r.modelName
}

func (r *GenkitRuntime) Available() bool {
	return r != nil &&
		r.available &&
		r.g != nil &&
		r.workoutDraftTool != nil
}

func (r *GenkitRuntime) GenerateValidation(ctx context.Context, prompt string) (*ValidationOutput, error) {
	if !r.Available() {
		return nil, ErrRuntimeUnavailable
	}

	output, _, err := genkit.GenerateData[ValidationOutput](ctx, r.g,
		ai.WithModelName(r.modelName),
		ai.WithOutputType(ValidationOutput{}),
		ai.WithPrompt(buildStructuredPrompt(prompt)),
	)
	if err != nil {
		return nil, fmt.Errorf("generate validation output: %w", err)
	}

	return output, nil
}

func (r *GenkitRuntime) StreamValidation(ctx context.Context, prompt string, onChunk func(string) error) (*StreamDone, error) {
	if !r.Available() {
		return nil, ErrRuntimeUnavailable
	}

	var builder strings.Builder
	resp, err := genkit.Generate(ctx, r.g,
		ai.WithModelName(r.modelName),
		ai.WithPrompt(buildStreamingPrompt(prompt)),
		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			delta := collectChunkText(chunk)
			if delta == "" {
				return nil
			}
			builder.WriteString(delta)
			return onChunk(delta)
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("stream validation output: %w", err)
	}

	text := strings.TrimSpace(resp.Text())
	if text == "" {
		text = strings.TrimSpace(builder.String())
	}

	return &StreamDone{
		Model: r.modelName,
		Text:  text,
	}, nil
}

func (r *GenkitRuntime) StreamChat(ctx context.Context, prompt string, history []RuntimeChatMessage, onChunk func(string) error) (*StreamDone, error) {
	if !r.Available() {
		return nil, ErrRuntimeUnavailable
	}

	traceStartedAt := time.Now()
	streamCtx, cancel := context.WithTimeout(ctx, chatStreamTimeout)
	defer cancel()

	var builder strings.Builder
	firstModelDelta := false
	snapshot := r.trainingSnapshotForChat(ctx)
	profile := r.trainingProfileForChat(ctx)
	opts := []ai.GenerateOption{
		ai.WithModelName(r.modelName),
		ai.WithTools(r.chatTools()...),
		ai.WithMaxTurns(chatMaxTurns),
		ai.WithMessages(buildChatMessages(history, snapshot, profile, r.dataReader != nil)...),
		ai.WithPrompt(prompt),
		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			delta := collectChunkText(chunk)
			if delta == "" {
				return nil
			}
			if !firstModelDelta {
				firstModelDelta = true
				// Trace marker: confirms when Gemini produced the first text delta.
				logAIChatTraceContext(ctx, "runtime_first_model_delta",
					"elapsed_ms", time.Since(traceStartedAt).Milliseconds(),
					"model", r.modelName,
					"history_messages", len(history),
					"request_id", request.GetRequestID(ctx),
				)
			}
			builder.WriteString(delta)
			return onChunk(delta)
		}),
	}

	// Trace marker: starts the outer chat generation, including any tool planning.
	logAIChatTraceContext(ctx, "runtime_generate_started",
		"model", r.modelName,
		"history_messages", len(history),
		"request_id", request.GetRequestID(ctx),
	)
	var resp *ai.ModelResponse
	for attempt := 1; ; attempt++ {
		builder.Reset()
		var err error
		resp, err = genkit.Generate(streamCtx, r.g, opts...)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(streamCtx.Err(), context.DeadlineExceeded) {
				return nil, fmt.Errorf("%w: %v", ErrGenerationTimeout, err)
			}
			return nil, fmt.Errorf("stream ai chat response: %w", err)
		}
		if attempt > chatEmptyResponseRetryLimit || !isEmptyChatModelResponse(resp, builder.String()) {
			break
		}
		// Gemini intermittently returns an empty candidate (for example on
		// MALFORMED_FUNCTION_CALL) when multiple tools are registered; an empty
		// reply is never a valid chat outcome, so retry the generation.
		slog.Warn("ai chat model returned an empty response with no tool activity; retrying",
			"model", r.modelName,
			"finish_reason", resp.FinishReason,
			"finish_message", resp.FinishMessage,
			"attempt", attempt,
			"request_id", request.GetRequestID(ctx),
		)
	}
	logAIChatTraceContext(ctx, "runtime_generate_finished",
		"elapsed_ms", time.Since(traceStartedAt).Milliseconds(),
		"model", r.modelName,
		"history_messages", len(history),
		"request_id", request.GetRequestID(ctx),
	)

	text := strings.TrimSpace(resp.Text())
	if text == "" {
		text = strings.TrimSpace(builder.String())
	}
	workoutDraft, err := extractWorkoutDraftFromHistory(resp.History())
	if err != nil {
		return nil, fmt.Errorf("extract workout draft from history: %w", err)
	}
	toolCalls := extractToolCallsFromHistory(resp.History())
	text = finalizeAssistantText(text, workoutDraft)

	return &StreamDone{
		Model:        r.modelName,
		Text:         text,
		WorkoutDraft: workoutDraft,
		ToolCalls:    toolCalls,
	}, nil
}

func isEmptyChatModelResponse(resp *ai.ModelResponse, streamedText string) bool {
	if resp == nil {
		return true
	}
	if strings.TrimSpace(resp.Text()) != "" || strings.TrimSpace(streamedText) != "" {
		return false
	}
	return len(extractToolCallsFromHistory(resp.History())) == 0
}

func (r *GenkitRuntime) chatTools() []ai.ToolRef {
	if r == nil || r.workoutDraftTool == nil {
		return nil
	}
	tools := []ai.ToolRef{r.workoutDraftTool}
	if r.getWorkoutsTool != nil {
		tools = append(tools, r.getWorkoutsTool)
	}
	if r.getExerciseStatsTool != nil {
		tools = append(tools, r.getExerciseStatsTool)
	}
	if r.updateProfileTool != nil {
		tools = append(tools, r.updateProfileTool)
	}
	return tools
}

func (r *GenkitRuntime) trainingSnapshotForChat(ctx context.Context) *TrainingSnapshot {
	if r == nil || r.dataReader == nil {
		return nil
	}
	userID, ok := user.Current(ctx)
	if !ok || strings.TrimSpace(userID) == "" {
		return nil
	}
	snapshot, err := r.dataReader.TrainingSnapshot(ctx, userID)
	if err != nil {
		slog.Warn("ai chat training snapshot omitted after reader error",
			"error", err,
			"request_id", request.GetRequestID(ctx),
		)
		return nil
	}
	return snapshot
}

func (r *GenkitRuntime) trainingProfileForChat(ctx context.Context) *TrainingProfile {
	if r == nil || r.dataReader == nil {
		return nil
	}
	userID, ok := user.Current(ctx)
	if !ok || strings.TrimSpace(userID) == "" {
		return nil
	}
	profile, err := r.dataReader.TrainingProfile(ctx, userID)
	if err != nil {
		slog.Warn("ai chat training profile omitted after reader error",
			"error", err,
			"request_id", request.GetRequestID(ctx),
		)
		return nil
	}
	return profile
}

func configuredAPIKeyEnvVar() string {
	if strings.TrimSpace(os.Getenv(geminiAPIKeyEnvVar)) != "" {
		return geminiAPIKeyEnvVar
	}
	if strings.TrimSpace(os.Getenv(googleAPIKeyEnvVar)) != "" {
		return googleAPIKeyEnvVar
	}
	return ""
}

func resolveModelName() string {
	if modelName := strings.TrimSpace(os.Getenv("GEMINI_MODEL")); modelName != "" {
		return modelName
	}
	return defaultModelName
}

func buildStructuredPrompt(prompt string) string {
	return fmt.Sprintf(`You are validating the phase-0 FitTrack AI chat architecture.

Return JSON with:
- "summary": one short sentence describing whether the architecture is viable for the authenticated viewer.
- "next_step": one short sentence naming the highest-value phase-1 implementation step.

Keep the response concise and grounded in this user validation prompt:
%s`, prompt)
}

func buildStreamingPrompt(prompt string) string {
	return fmt.Sprintf(`You are validating the FitTrack phase-0 streaming path.

Respond in 2 short sentences.
- Sentence 1 should confirm that authenticated fetch-based SSE is viable inside the existing Go API.
- Sentence 2 should name the most important server-side concern to address next.

User validation prompt:
%s`, prompt)
}

func buildChatMessages(history []RuntimeChatMessage, snapshot *TrainingSnapshot, profile *TrainingProfile, dataToolsEnabled bool) []*ai.Message {
	messages := []*ai.Message{
		ai.NewSystemMessage(ai.NewTextPart(buildChatSystemPrompt(snapshot, profile, time.Now(), dataToolsEnabled))),
	}

	for _, message := range history {
		text := strings.TrimSpace(message.Text)
		if text == "" {
			continue
		}
		switch message.Role {
		case roleAssistant:
			messages = append(messages, ai.NewModelMessage(ai.NewTextPart(text)))
		default:
			messages = append(messages, ai.NewUserMessage(ai.NewTextPart(text)))
		}
	}

	return messages
}

func buildChatSystemPrompt(snapshot *TrainingSnapshot, profile *TrainingProfile, now time.Time, dataToolsEnabled bool) string {
	personalDataRule := "- Never guess or invent personal workout history. Say you do not have workout data available in this chat when asked about personal training history. Do not call data tools for general fitness knowledge."
	if dataToolsEnabled {
		personalDataRule = fmt.Sprintf(`- For questions about the user's logged workouts or personal training history, call the %s tool. Never guess or invent workout history; if no data exists, say so. Do not call data tools for general fitness knowledge.
- Default to %s for personal workout-history questions; use %s only for all-time bests, PRs, estimated 1RM, or long-range single-exercise trends.
- The data tools and the training snapshot exist only to answer questions about past training or to supply recentPerformance for a requested workout draft. When the user asks you to build a workout, do not call data tools unless the user explicitly references past training, such as "based on last week", "like last time", or "heavier than last time". Always respond with either a follow-up question or a %s tool call, never an empty reply. The snapshot and logged history never satisfy the workout-building inputs below — in particular, they say nothing about current injury status, so still ask about injuries when that is missing.`, getWorkoutsToolName, getWorkoutsToolName, getExerciseStatsToolName, workoutDraftToolName)
	}
	currentDateSection := fmt.Sprintf("Current date: %s.", now.Format("2006-01-02"))
	snapshotSection := buildTrainingSnapshotPromptSection(snapshot)
	profileSection := buildTrainingProfilePromptSection(profile)

	return fmt.Sprintf(`You are FitTrack's in-app training assistant.

Rules:
- Stay focused on fitness, training, recovery, exercise selection, and how to use FitTrack.
- Keep answers concise, practical, and safe.
%s
- Call the %s tool only for durable training facts the user stated, such as usual equipment, usual location, primary goal, experience level, exercises to avoid, or movement limitations. Do not call it for one-off session details like "today only" equipment, location, duration, or constraints.
- When you call the %s tool, tell the user what you saved. If the user asks you to forget a profile fact, call the tool with that field cleared and tell them what was cleared. For profile array fields, send the complete updated list because arrays replace the saved list. Never say you saved, remembered, or updated profile facts unless you called the %s tool in this same turn; if you are not calling the tool, do not claim any memory.

%s
%s
%s

When the user wants you to build a workout:
- Review the visible conversation first and reason about which workout inputs are already confirmed versus still missing.
- First separate confirmed inputs from missing inputs. MVP-ready inputs are workout focus, session duration, enough equipment or workout context to choose feasible exercises, and injury status.
- Ask at most %d short, focused follow-up questions at a time for the missing MVP-ready inputs.
- Ask for fitness level when it is missing because it improves the baseline and weight assumptions, but do not treat it as a hard blocker once the MVP-ready inputs are present.
- Equipment is optional for mobility, rehab, prehab, stretching, or warm-up requests. Resistance bands, foam rollers, sticks, and similar tools can add challenge, support, regression, progression, or convenience, but they are not required before generating.
- For normal strength, hypertrophy, endurance, cardio, or general fitness workouts, do not call the %s tool until the user has provided equipment, training location, or space constraints. If this is missing, ask where they will train and what equipment they have.
- Treat the user's stated equipment, training location, or space constraints as the available context for the draft. Use only that context unless the user explicitly mentions more. Do not ask what other equipment they have unless the requested workout is unsafe, contradictory, or not reasonably buildable with the stated constraints. Do not assume unmentioned accessories or equipment, such as a bench, rack, cable, or machine.
- Treat user profile values as defaults for workout drafts. The user's current message always overrides the profile. Do not re-ask for known profile equipment, location, or movement limitations; if the profile says movement limitations are none stated, treat injury status as known none, use injuries="none", and do not ask about injuries before drafting — this overrides the injury-question rules below. Briefly state assumptions when using profile defaults, such as "Using your usual home dumbbell setup — say the word if today's different."
- If injury status is missing and no profile movement limitation default is available, ask once before generating. Do not infer "none" from silence in the initial request, even when the rest of the workout request is clear.
- Use injuries="none" only when the user explicitly says they have no injuries or when you already asked about injuries and the user continues without answering.
- When the user answers a follow-up, combine that answer with the earlier visible workout request. If your previous message only asked about injuries and the user now confirms no injuries, reuse the earlier focus, duration, equipment, and location details instead of asking them to repeat those details.
- After one follow-up answer, if workout focus, session duration, equipment or location context, and usable injury details are present, call the %s tool with conservative assumptions. Do not ask optional questions about fitness level, preferred exercises, or other movements to avoid unless the pain description is unclear, severe, includes red flags, or the stated constraints cannot support a safe, reasonable draft.
- FitTrack creates one structured workout draft at a time, not full weekly splits, multi-day programs, or bundled plans. For a multi-day request, ask the user to choose one day, workout, or session to build first and ask only for missing MVP-ready inputs for that single session.
- If the user follows up after a multi-day request with one day or session, such as "day one" or "start with upper body", treat that as a valid single-session scope. Once that follow-up also includes MVP-ready inputs, call the %s tool immediately instead of asking them to confirm the scope again.
- Do not ask scheduling, frequency, or future-date questions in the normal MVP flow. If the user does not specify a date, the draft tool will default the workout date to today.
- Do not list specific exercises, sets, or reps in plain text before the %s tool runs.
- As soon as you have the MVP-ready inputs, call the %s tool immediately.
- The %s tool is the only way to produce a structured workout draft that matches FitTrack's workout contract.
- After the tool runs, keep any follow-up text to a short summary and do not repeat the exercise list in plain text.
- If the user adds a new pain, injury, or movement limitation after a workout draft, treat it as new safety context before revising. When the new limitation is vague, ask one focused follow-up about the painful movements, ranges, or triggers before calling the %s tool again. If the limitation is specific enough to revise safely, call the %s tool right away.

Examples:
- If the user says "I want a chest workout," ask only for the missing requirements instead of drafting exercises.
- If the user says "I'd like a fitness plan" and later says "Upper body. 30 minutes. No injuries.", ask where they will train and what equipment they have before calling the %s tool.
- If the user gives focus, duration, and equipment but does not mention injuries, ask about injuries before calling the %s tool.
- If the user says "Full gym, 45 minutes, hypertrophy pull day, no injuries," call the %s tool right away even if fitness level is unknown.
- If the user says they have knee pain and want a lower-body workout, then answers "mild front-of-knee discomfort, no sharp pain, 35 minutes, dumbbells and machines," call the %s tool. Do not ask for fitness level first.
- If the user says they recently strained a shoulder and overhead pressing bothers it, then gives 30 minutes, dumbbells, and no sharp pain, call the %s tool with shoulder-friendly dumbbell-only choices that do not assume a bench. Do not ask whether any other movements bother them unless the pain description is unclear or severe.
- If the user says "45-minute back workout, cables only, no injuries," call the workout draft tool and choose cable-only movements instead of asking what other equipment they have.
- If the user first asks for a 4-day split, say FitTrack builds one workout at a time and ask them to choose one day or session to start. If they then say "Let's start with day one as an upper-body workout. No injuries, full gym, 45 minutes," call the %s tool for that upper-body session.
- If the user says "swap anything that bothers my knee/elbow/shoulder/back/wrist" after a draft, ask which movements, ranges, or exercise patterns bother that body part before revising.
- If the user asks to swap or revise a generated workout later, gather only the extra details needed for the revision and stay concise.`, personalDataRule, updateTrainingProfileToolName, updateTrainingProfileToolName, updateTrainingProfileToolName, currentDateSection, snapshotSection, profileSection, workoutChatFollowUpQuestionCeiling, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName)
}

func buildTrainingSnapshotPromptSection(snapshot *TrainingSnapshot) string {
	if snapshot == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("\nUser training snapshot:\n")
	if strings.TrimSpace(snapshot.LastWorkoutDate) != "" {
		builder.WriteString(fmt.Sprintf("- Last workout: %s\n", snapshot.LastWorkoutDate))
	} else {
		builder.WriteString("- Last workout: none logged\n")
	}
	builder.WriteString(fmt.Sprintf("- Workouts in last 30 days: %d\n", snapshot.WorkoutsLast30D))
	if len(snapshot.TopExercises) > 0 {
		builder.WriteString(fmt.Sprintf("- Most frequent exercises: %s\n", strings.Join(snapshot.TopExercises, ", ")))
	}
	return strings.TrimRight(builder.String(), "\n")
}

func buildTrainingProfilePromptSection(profile *TrainingProfile) string {
	if !hasTrainingProfileContent(profile) {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("\nUser training profile:\n")
	if strings.TrimSpace(profile.PrimaryGoal) != "" {
		builder.WriteString(fmt.Sprintf("- Goal: %s\n", profile.PrimaryGoal))
	}
	if strings.TrimSpace(profile.ExperienceLevel) != "" {
		builder.WriteString(fmt.Sprintf("- Experience: %s\n", profile.ExperienceLevel))
	}
	if profile.PreferredSessionDurationMinutes > 0 {
		builder.WriteString(fmt.Sprintf("- Preferred duration: %d minutes\n", profile.PreferredSessionDurationMinutes))
	}
	if strings.TrimSpace(profile.UsualTrainingLocation) != "" {
		builder.WriteString(fmt.Sprintf("- Usual location: %s\n", profile.UsualTrainingLocation))
	}
	if len(profile.AvailableEquipment) > 0 {
		builder.WriteString(fmt.Sprintf("- Available equipment: %s\n", strings.Join(profile.AvailableEquipment, ", ")))
	}
	if len(profile.AvoidedExercises) > 0 {
		builder.WriteString(fmt.Sprintf("- Avoided exercises: %s\n", strings.Join(profile.AvoidedExercises, ", ")))
	}
	if len(profile.MovementLimitations) > 0 {
		builder.WriteString(fmt.Sprintf("- Movement limitations: %s\n", strings.Join(profile.MovementLimitations, ", ")))
	} else {
		builder.WriteString("- Movement limitations: none stated\n")
	}
	return strings.TrimRight(builder.String(), "\n")
}

func collectChunkText(chunk *ai.ModelResponseChunk) string {
	var builder strings.Builder
	for _, part := range chunk.Content {
		builder.WriteString(part.Text)
	}
	return builder.String()
}
