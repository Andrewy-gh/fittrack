package aichat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"

	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
)

const (
	geminiAPIKeyEnvVar     = "GEMINI_API_KEY"
	googleAPIKeyEnvVar     = "GOOGLE_API_KEY"
	chatStreamTimeout      = 45 * time.Second
	chatMaxTurns           = 6
	activeFeaturesToolName = "fittrack.list_active_features"
)

var genkitInit = func(ctx context.Context, opts ...genkit.GenkitOption) *genkit.Genkit {
	return genkit.Init(ctx, opts...)
}

type featureAccessReader interface {
	ListCurrentUserAccess(ctx context.Context) ([]featureaccess.FeatureAccessGrant, error)
}

type GenkitRuntime struct {
	available          bool
	modelName          string
	g                  *genkit.Genkit
	activeFeaturesTool ai.Tool
	workoutDraftTool   ai.Tool
}

func NewGenkitRuntime(ctx context.Context, featureAccess featureAccessReader) *GenkitRuntime {
	modelName := resolveModelName()
	runtime := &GenkitRuntime{
		modelName: modelName,
	}

	if configuredAPIKeyEnvVar() == "" {
		return runtime
	}

	g, activeFeaturesTool, workoutDraftTool, ok := activateGenkitRuntime(ctx, modelName, featureAccess)
	if !ok {
		return runtime
	}

	runtime.g = g
	runtime.activeFeaturesTool = activeFeaturesTool
	runtime.workoutDraftTool = workoutDraftTool
	runtime.available = true

	return runtime
}

func activateGenkitRuntime(ctx context.Context, modelName string, featureAccess featureAccessReader) (_ *genkit.Genkit, _ ai.Tool, _ ai.Tool, ok bool) {
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

	activeFeaturesTool := genkit.DefineTool(g, activeFeaturesToolName,
		"Lists the active FitTrack feature keys for the authenticated viewer.",
		func(ctx *ai.ToolContext, _ struct{}) (*featureSnapshot, error) {
			guard := toolGuardFromContext(ctx)
			if guard == nil {
				return listActiveFeaturesSnapshot(ctx, featureAccess)
			}
			return guard.listActiveFeatures(ctx, featureAccess)
		},
	)

	workoutDraftTool := defineWorkoutDraftTool(g, modelName)

	return g, activeFeaturesTool, workoutDraftTool, true
}

func (r *GenkitRuntime) ModelName() string {
	return r.modelName
}

func (r *GenkitRuntime) Available() bool {
	return r != nil &&
		r.available &&
		r.g != nil &&
		r.activeFeaturesTool != nil &&
		r.workoutDraftTool != nil
}

func (r *GenkitRuntime) GenerateValidation(ctx context.Context, prompt string) (*ValidationOutput, error) {
	if !r.Available() {
		return nil, ErrRuntimeUnavailable
	}

	output, _, err := genkit.GenerateData[ValidationOutput](withToolGuard(ctx), r.g,
		ai.WithModelName(r.modelName),
		ai.WithOutputType(ValidationOutput{}),
		ai.WithTools(r.activeFeaturesTool),
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
	resp, err := genkit.Generate(withToolGuard(ctx), r.g,
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

	streamCtx, cancel := context.WithTimeout(ctx, chatStreamTimeout)
	defer cancel()
	streamCtx = withToolGuard(streamCtx)

	var builder strings.Builder
	opts := []ai.GenerateOption{
		ai.WithModelName(r.modelName),
		ai.WithTools(r.activeFeaturesTool, r.workoutDraftTool),
		ai.WithMaxTurns(chatMaxTurns),
		ai.WithMessages(buildChatMessages(history)...),
		ai.WithPrompt(prompt),
		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			delta := collectChunkText(chunk)
			if delta == "" {
				return nil
			}
			builder.WriteString(delta)
			return onChunk(delta)
		}),
	}

	resp, err := genkit.Generate(streamCtx, r.g, opts...)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(streamCtx.Err(), context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: %v", ErrGenerationTimeout, err)
		}
		return nil, fmt.Errorf("stream ai chat response: %w", err)
	}

	text := strings.TrimSpace(resp.Text())
	if text == "" {
		text = strings.TrimSpace(builder.String())
	}
	workoutDraft, err := extractWorkoutDraftFromHistory(resp.History())
	if err != nil {
		return nil, fmt.Errorf("extract workout draft from history: %w", err)
	}
	text = finalizeAssistantText(text, workoutDraft)

	return &StreamDone{
		Model:        r.modelName,
		Text:         text,
		WorkoutDraft: workoutDraft,
	}, nil
}

func listActiveFeaturesSnapshot(ctx context.Context, featureAccess featureAccessReader) (*featureSnapshot, error) {
	grants, err := featureAccess.ListCurrentUserAccess(ctx)
	if err != nil {
		return nil, err
	}

	keys := make([]string, 0, len(grants))
	for _, grant := range grants {
		keys = append(keys, grant.FeatureKey)
	}
	sort.Strings(keys)

	return &featureSnapshot{FeatureKeys: keys}, nil
}

type featureSnapshot struct {
	FeatureKeys []string `json:"feature_keys"`
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

You must call the "%s" tool exactly once before answering.

Return JSON with:
- "summary": one short sentence describing whether the architecture is viable for the authenticated viewer.
- "next_step": one short sentence naming the highest-value phase-1 implementation step.

Keep the response concise and grounded in the tool result plus this user validation prompt:
%s`, activeFeaturesToolName, prompt)
}

func buildStreamingPrompt(prompt string) string {
	return fmt.Sprintf(`You are validating the FitTrack phase-0 streaming path.

Respond in 2 short sentences.
- Sentence 1 should confirm that authenticated fetch-based SSE is viable inside the existing Go API.
- Sentence 2 should name the most important server-side concern to address next.

User validation prompt:
%s`, prompt)
}

func buildChatMessages(history []RuntimeChatMessage) []*ai.Message {
	messages := []*ai.Message{
		ai.NewSystemMessage(ai.NewTextPart(buildChatSystemPrompt())),
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

func buildChatSystemPrompt() string {
	return fmt.Sprintf(`You are FitTrack's in-app training assistant.

Rules:
- Stay focused on fitness, training, recovery, exercise selection, and how to use FitTrack.
- Keep answers concise, practical, and safe.
- Base your response on the visible conversation only. Do not invent personal history or workout data.
- If feature access is relevant, you may use the %s tool.

When the user wants you to build a workout:
- Review the visible conversation first and reason about which workout inputs are already confirmed versus still missing.
- First separate confirmed inputs from missing inputs. MVP-ready inputs are workout focus, session duration, enough equipment or workout context to choose feasible exercises, and injury status.
- Ask at most %d short, focused follow-up questions at a time for the missing MVP-ready inputs.
- Ask for fitness level when it is missing because it improves the baseline and weight assumptions, but do not treat it as a hard blocker once the MVP-ready inputs are present.
- If injury status is missing, ask once. If you already asked about injuries and the user continues without answering, assume injuries are "none" and proceed.
- Do not ask scheduling, frequency, or future-date questions in the normal MVP flow. If the user does not specify a date, the draft tool will default the workout date to today.
- Do not list specific exercises, sets, or reps in plain text before the %s tool runs.
- As soon as you have the MVP-ready inputs, call the %s tool immediately.
- The %s tool is the only way to produce a structured workout draft that matches FitTrack's workout contract.
- After the tool runs, keep any follow-up text to a short summary and do not repeat the exercise list in plain text.

Examples:
- If the user says "I want a chest workout," ask only for the missing requirements instead of drafting exercises.
- If the user says "Full gym, 45 minutes, hypertrophy pull day, no injuries," call the %s tool right away even if fitness level is unknown.
- If the user asks to swap or revise a generated workout later, gather only the extra details needed for the revision and stay concise.`, activeFeaturesToolName, workoutChatFollowUpQuestionCeiling, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName, workoutDraftToolName)
}

func collectChunkText(chunk *ai.ModelResponseChunk) string {
	var builder strings.Builder
	for _, part := range chunk.Content {
		builder.WriteString(part.Text)
	}
	return builder.String()
}
