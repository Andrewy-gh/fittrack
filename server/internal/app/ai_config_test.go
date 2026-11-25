package app

import (
	"context"
	"testing"

	"github.com/Andrewy-gh/fittrack/server/internal/aichat"
	"github.com/Andrewy-gh/fittrack/server/internal/config"
)

func TestLoadedAIChatSnapshotControlsOfflineRuntime(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://localhost/test")
	t.Setenv("PROJECT_ID", "synthetic")
	t.Setenv("ENVIRONMENT", "development")
	t.Setenv("RLS_ENFORCEMENT_REQUIRED", "")
	t.Setenv("GEMINI_API_KEY", "")
	t.Setenv("GOOGLE_API_KEY", "")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("GEMINI_MODEL", " before/model ")
	t.Setenv("AI_CHAT_PRIMARY_MODEL", "")
	t.Setenv("AI_CHAT_FALLBACK", "")

	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("GEMINI_MODEL", "after/model")
	t.Setenv("AI_CHAT_PRIMARY_MODEL", config.LunaModel)
	t.Setenv("AI_CHAT_FALLBACK", "gemini")
	runtime := aichat.NewGenkitRuntime(context.Background(), nil, cfg.AIChat)
	if runtime.ModelName() != "before/model" || runtime.Available() {
		t.Fatalf("runtime did not use startup model snapshot: model=%q available=%t", runtime.ModelName(), runtime.Available())
	}
	if cfg.AIChat.GeminiFallback() {
		t.Fatal("fallback changed after startup configuration was loaded")
	}
	fresh := aichat.NewGenkitRuntime(context.Background(), nil, config.LoadAIChat())
	if fresh.ModelName() != config.LunaModel {
		t.Fatal("later CLI-style parse did not see the model override")
	}
	// Callers constructing Config directly still get the historical default.
	var zero config.Config
	defaultRuntime := aichat.NewGenkitRuntime(context.Background(), nil, zero.AIChat)
	if defaultRuntime.ModelName() != config.DefaultGeminiModel || defaultRuntime.Available() {
		t.Fatal("zero configuration did not retain the unavailable default runtime")
	}
}
