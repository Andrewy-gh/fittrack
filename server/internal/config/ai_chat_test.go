package config

import "testing"

func TestLoadAIChatModelSelection(t *testing.T) {
	for _, tc := range []struct {
		name, gemini, primary, fallback, wantGemini, wantPrimary string
		wantFallback                                             bool
	}{
		{"defaults", "", "", "", DefaultGeminiModel, DefaultGeminiModel, false},
		{"blank model", " \t", "other", "GEMINI", DefaultGeminiModel, DefaultGeminiModel, false},
		{"trim gemini", " custom/model ", "unknown", "gemini", "custom/model", "custom/model", true},
		{"luna with fallback", " custom/model ", "  openai/gpt-6-luna ", "gemini", "custom/model", LunaModel, true},
		{"fallback is exact", "", LunaModel, " gemini ", DefaultGeminiModel, LunaModel, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GEMINI_MODEL", tc.gemini)
			t.Setenv("AI_CHAT_PRIMARY_MODEL", tc.primary)
			t.Setenv("AI_CHAT_FALLBACK", tc.fallback)
			got := LoadAIChat()
			if got.GeminiModel() != tc.wantGemini || got.PrimaryModel() != tc.wantPrimary || got.GeminiFallback() != tc.wantFallback {
				t.Fatalf("unexpected parsed model selection: gemini=%q primary=%q fallback=%v", got.GeminiModel(), got.PrimaryModel(), got.GeminiFallback())
			}
			// A runtime snapshot must not change when later overrides are applied.
			t.Setenv("GEMINI_MODEL", "later")
			if got.GeminiModel() != tc.wantGemini || LoadAIChat().GeminiModel() != "later" {
				t.Fatal("model snapshot or subsequent boundary parse is wrong")
			}
		})
	}
}

func TestAIChatTraceEnabled(t *testing.T) {
	for _, tc := range []struct {
		value string
		want  bool
	}{
		{"", false}, {"1", true}, {" TRUE ", true}, {"Yes", true}, {"on", true},
		{"0", false}, {"false", false}, {"off", false}, {"no", false}, {"invalid", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			t.Setenv("AI_CHAT_TRACE_LOGS", tc.value)
			if got := AIChatTraceEnabled(); got != tc.want {
				t.Fatalf("trace enabled = %v, want %v", got, tc.want)
			}
		})
	}
}
