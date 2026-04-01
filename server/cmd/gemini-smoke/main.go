package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/joho/godotenv"
)

const (
	geminiAPIKeyEnvVar = "GEMINI_API_KEY"
	googleAPIKeyEnvVar = "GOOGLE_API_KEY"
	defaultGeminiModel = "googleai/gemini-2.5-flash"
	smokePrompt        = "Reply with one short sentence confirming the FitTrack Genkit smoke test is working."
	smokeTimeout       = 20 * time.Second
)

func main() {
	if err := loadLocalEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to load local env files: %v\n", err)
		os.Exit(1)
	}

	if configuredAPIKeyEnvVar() == "" {
		fmt.Fprintf(os.Stderr, "%s or %s must be set. Put one in your shell or in server/.env, server/.setenv.sh, or server/setenv.sh.\n", geminiAPIKeyEnvVar, googleAPIKeyEnvVar)
		os.Exit(1)
	}

	modelName := getModelName()

	ctx, cancel := context.WithTimeout(context.Background(), smokeTimeout)
	defer cancel()
	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel(modelName),
	)

	resp, err := genkit.Generate(ctx, g,
		ai.WithMessages(ai.NewUserMessage(ai.NewTextPart(smokePrompt))),
		ai.WithModelName(modelName),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, formatRunError(err, modelName))
		os.Exit(1)
	}

	text := formatResponse(resp.Text())
	if text == "" {
		fmt.Fprintln(os.Stderr, "genkit gemini smoke test returned an empty response")
		os.Exit(1)
	}

	fmt.Println(text)
}

func loadLocalEnv() error {
	files := existingEnvFiles(localEnvFiles()...)
	if len(files) == 0 {
		return nil
	}

	if err := godotenv.Load(files...); err != nil {
		return fmt.Errorf("load %v: %w", files, err)
	}

	return nil
}

func localEnvFiles() []string {
	return []string{".env", ".setenv.sh", "setenv.sh"}
}

func existingEnvFiles(paths ...string) []string {
	files := make([]string, 0, len(paths))
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			files = append(files, path)
		}
	}
	return files
}

func formatResponse(text string) string {
	const maxLen = 160

	cleaned := strings.Join(strings.Fields(text), " ")
	if len(cleaned) <= maxLen {
		return cleaned
	}

	return strings.TrimSpace(cleaned[:maxLen-3]) + "..."
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

func getModelName() string {
	if model := strings.TrimSpace(os.Getenv("GEMINI_MODEL")); model != "" {
		return model
	}
	return defaultGeminiModel
}

func formatRunError(err error, modelName string) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Sprintf("genkit gemini smoke test timed out after %s for model %s. Check network access to Gemini and try again.", smokeTimeout, modelName)
	}

	raw := strings.TrimSpace(err.Error())
	if isQuotaOr429Error(raw) {
		return quotaOr429ErrorMessage(raw, modelName)
	}

	return fmt.Sprintf("genkit gemini smoke test failed for model %s: %s", modelName, raw)
}

func isQuotaOr429Error(raw string) bool {
	lower := strings.ToLower(raw)
	return strings.Contains(lower, "error 429") ||
		strings.Contains(lower, "resource_exhausted") ||
		strings.Contains(lower, "quota exceeded")
}

func quotaOr429ErrorMessage(raw string, modelName string) string {
	lines := []string{
		fmt.Sprintf("genkit gemini smoke test failed for model %s.", modelName),
		"",
		"Google accepted the API key, but rejected the request with a quota or project-tier error (HTTP 429 / RESOURCE_EXHAUSTED).",
		"",
		"Next checks:",
		"- In Google AI Studio, confirm this API key belongs to the project you intended to use.",
		"- Verify billing/tier for that project. Free-tier projects can return quota limit 0 for some models.",
		"- Verify the selected model is available to that project and region.",
	}

	if modelName != defaultGeminiModel {
		lines = append(lines, fmt.Sprintf("- Retry with the default smoke-test model: `$env:GEMINI_MODEL=\"%s\"; go run ./cmd/gemini-smoke`", defaultGeminiModel))
	} else {
		lines = append(lines, "- Retry the same command after checking the project's Gemini API quotas in AI Studio.")
	}

	lines = append(lines,
		"",
		"Raw error:",
		fmt.Sprintf("- %s", raw),
	)

	return strings.Join(lines, "\n")
}
