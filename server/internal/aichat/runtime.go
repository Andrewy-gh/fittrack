package aichat

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"

	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
)

const (
	geminiAPIKeyEnvVar       = "GEMINI_API_KEY"
	googleAPIKeyEnvVar       = "GOOGLE_API_KEY"
	debugStreamDelayEnvVar   = "AI_CHAT_DEBUG_STREAM_DELAY_MS"
	debugForceRecoveryEnvVar = "AI_CHAT_DEBUG_FORCE_RECOVERY_AFTER_CHUNKS"
	chatStreamTimeout        = 45 * time.Second
	debugStreamChunkRuneSize = 12
	activeFeaturesToolName = "fittrack.list_active_features"
)

var genkitInit = func(ctx context.Context, opts ...genkit.GenkitOption) *genkit.Genkit {
	return genkit.Init(ctx, opts...)
}

var sleepWithContext = func(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type featureAccessReader interface {
	ListCurrentUserAccess(ctx context.Context) ([]featureaccess.FeatureAccessGrant, error)
}

type GenkitRuntime struct {
	available bool
	modelName string
	g         *genkit.Genkit
	tool      ai.Tool
}

type streamDebugState struct {
	delay                    time.Duration
	forceRecoveryAfterChunks int
	emittedChunks            int
}

type foregroundStreamDebugContextKey struct{}

func NewGenkitRuntime(ctx context.Context, featureAccess featureAccessReader) *GenkitRuntime {
	modelName := resolveModelName()
	runtime := &GenkitRuntime{
		modelName: modelName,
	}

	if configuredAPIKeyEnvVar() == "" {
		return runtime
	}

	g, tool, ok := activateGenkitRuntime(ctx, modelName, featureAccess)
	if !ok {
		return runtime
	}

	runtime.g = g
	runtime.tool = tool
	runtime.available = true

	return runtime
}

func activateGenkitRuntime(ctx context.Context, modelName string, featureAccess featureAccessReader) (_ *genkit.Genkit, _ ai.Tool, ok bool) {
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

	tool := genkit.DefineTool(g, activeFeaturesToolName,
		"Lists the active FitTrack feature keys for the authenticated viewer.",
		func(ctx *ai.ToolContext, _ struct{}) (*featureSnapshot, error) {
			guard := toolGuardFromContext(ctx)
			if guard == nil {
				return listActiveFeaturesSnapshot(ctx, featureAccess)
			}
			return guard.listActiveFeatures(ctx, featureAccess)
		},
	)

	return g, tool, true
}

func (r *GenkitRuntime) ModelName() string {
	return r.modelName
}

func (r *GenkitRuntime) Available() bool {
	return r != nil && r.available && r.g != nil && r.tool != nil
}

func (r *GenkitRuntime) GenerateValidation(ctx context.Context, prompt string) (*ValidationOutput, error) {
	if !r.Available() {
		return nil, ErrRuntimeUnavailable
	}

	output, _, err := genkit.GenerateData[ValidationOutput](withToolGuard(ctx), r.g,
		ai.WithModelName(r.modelName),
		ai.WithOutputType(ValidationOutput{}),
		ai.WithTools(r.tool),
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
	debugState := newStreamDebugState(false)
	resp, err := genkit.Generate(withToolGuard(ctx), r.g,
		ai.WithModelName(r.modelName),
		ai.WithPrompt(buildStreamingPrompt(prompt)),
		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			delta := collectChunkText(chunk)
			if delta == "" {
				return nil
			}
			builder.WriteString(delta)
			return emitStreamText(ctx, delta, onChunk, debugState)
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
	debugState := newStreamDebugState(foregroundStreamDebugEnabled(streamCtx))
	opts := []ai.GenerateOption{
		ai.WithModelName(r.modelName),
		ai.WithTools(r.tool),
		ai.WithMessages(buildChatMessages(history)...),
		ai.WithPrompt(prompt),
		ai.WithStreaming(func(ctx context.Context, chunk *ai.ModelResponseChunk) error {
			delta := collectChunkText(chunk)
			if delta == "" {
				return nil
			}
			builder.WriteString(delta)
			return emitStreamText(ctx, delta, onChunk, debugState)
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

	return &StreamDone{
		Model: r.modelName,
		Text:  text,
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

func debugStreamDelay() time.Duration {
	raw := strings.TrimSpace(os.Getenv(debugStreamDelayEnvVar))
	if raw == "" {
		return 0
	}

	delayMs, err := strconv.Atoi(raw)
	if err != nil || delayMs <= 0 {
		return 0
	}

	return time.Duration(delayMs) * time.Millisecond
}

func debugForceRecoveryAfterChunks() int {
	raw := strings.TrimSpace(os.Getenv(debugForceRecoveryEnvVar))
	if raw == "" {
		return 0
	}

	chunkCount, err := strconv.Atoi(raw)
	if err != nil || chunkCount <= 0 {
		return 0
	}

	return chunkCount
}

func newStreamDebugState(enableForcedRecovery bool) *streamDebugState {
	state := &streamDebugState{
		delay: debugStreamDelay(),
	}
	if enableForcedRecovery {
		state.forceRecoveryAfterChunks = debugForceRecoveryAfterChunks()
	}
	return state
}

func withForegroundStreamDebug(ctx context.Context, enabled bool) context.Context {
	return context.WithValue(ctx, foregroundStreamDebugContextKey{}, enabled)
}

func foregroundStreamDebugEnabled(ctx context.Context) bool {
	enabled, _ := ctx.Value(foregroundStreamDebugContextKey{}).(bool)
	return enabled
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
- If feature access is relevant, you may use the %s tool.`, activeFeaturesToolName)
}

func collectChunkText(chunk *ai.ModelResponseChunk) string {
	var builder strings.Builder
	for _, part := range chunk.Content {
		builder.WriteString(part.Text)
	}
	return builder.String()
}

func emitStreamText(ctx context.Context, text string, onChunk func(string) error, debugState *streamDebugState) error {
	if onChunk == nil || text == "" {
		return nil
	}

	delay := time.Duration(0)
	if debugState != nil {
		delay = debugState.delay
	}
	if delay <= 0 {
		if err := onChunk(text); err != nil {
			return err
		}
		return maybeForceDebugRecovery(debugState)
	}

	for index, part := range splitStreamText(text, debugStreamChunkRuneSize) {
		if index > 0 {
			if err := sleepWithContext(ctx, delay); err != nil {
				return err
			}
		}
		if err := onChunk(part); err != nil {
			return err
		}
		if err := maybeForceDebugRecovery(debugState); err != nil {
			return err
		}
	}

	return nil
}

func maybeForceDebugRecovery(debugState *streamDebugState) error {
	if debugState == nil {
		return nil
	}

	debugState.emittedChunks++
	if debugState.forceRecoveryAfterChunks > 0 && debugState.emittedChunks >= debugState.forceRecoveryAfterChunks {
		return ErrStreamDisconnected
	}

	return nil
}

func splitStreamText(text string, chunkSize int) []string {
	if text == "" {
		return nil
	}
	if chunkSize <= 0 {
		return []string{text}
	}

	runes := []rune(text)
	chunks := make([]string, 0, (len(runes)+chunkSize-1)/chunkSize)
	for start := 0; start < len(runes); start += chunkSize {
		end := start + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[start:end]))
	}

	return chunks
}
