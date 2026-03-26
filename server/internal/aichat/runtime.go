package aichat

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"

	"github.com/Andrewy-gh/fittrack/server/internal/featureaccess"
)

const googleAPIKeyEnvVar = "GOOGLE_API_KEY"

type featureAccessReader interface {
	ListCurrentUserAccess(ctx context.Context) ([]featureaccess.FeatureAccessGrant, error)
}

type GenkitRuntime struct {
	available bool
	modelName string
	g         *genkit.Genkit
	tool      ai.Tool
}

func NewGenkitRuntime(ctx context.Context, featureAccess featureAccessReader) *GenkitRuntime {
	modelName := resolveModelName()
	runtime := &GenkitRuntime{
		modelName: modelName,
	}

	if strings.TrimSpace(os.Getenv(googleAPIKeyEnvVar)) == "" {
		return runtime
	}

	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel(modelName),
	)

	runtime.g = g
	runtime.tool = genkit.DefineTool(g, "fittrack/listActiveFeatures",
		"Lists the active FitTrack feature keys for the authenticated viewer.",
		func(ctx *ai.ToolContext, _ struct{}) (*featureSnapshot, error) {
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
		},
	)
	runtime.available = true

	return runtime
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

	output, _, err := genkit.GenerateData[ValidationOutput](ctx, r.g,
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

type featureSnapshot struct {
	FeatureKeys []string `json:"feature_keys"`
}

func resolveModelName() string {
	if modelName := strings.TrimSpace(os.Getenv("GEMINI_MODEL")); modelName != "" {
		return modelName
	}
	return defaultModelName
}

func buildStructuredPrompt(prompt string) string {
	return fmt.Sprintf(`You are validating the phase-0 FitTrack AI chat architecture.

You must call the "fittrack/listActiveFeatures" tool exactly once before answering.

Return JSON with:
- "summary": one short sentence describing whether the architecture is viable for the authenticated viewer.
- "next_step": one short sentence naming the highest-value phase-1 implementation step.

Keep the response concise and grounded in the tool result plus this user validation prompt:
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

func collectChunkText(chunk *ai.ModelResponseChunk) string {
	var builder strings.Builder
	for _, part := range chunk.Content {
		builder.WriteString(part.Text)
	}
	return builder.String()
}
