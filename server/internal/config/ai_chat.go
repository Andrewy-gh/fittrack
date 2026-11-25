package config

import (
	"os"
	"strings"
)

const (
	// DefaultGeminiModel is used when GEMINI_MODEL is blank.
	DefaultGeminiModel = "googleai/gemini-2.5-flash"
	// LunaModel is the recognized Luna primary model selector.
	LunaModel = "openai/gpt-6-luna"
)

// AIChat contains normalized application-owned model selection. Provider credentials
// remain with their provider integrations and are never stored here.
type AIChat struct {
	geminiModel    string
	primaryModel   string
	geminiFallback bool
}

// LoadAIChat parses model selection at the point a runtime is constructed.
// Unknown primary and fallback values retain the historical disabled/default behavior.
func LoadAIChat() AIChat {
	gemini := strings.TrimSpace(os.Getenv("GEMINI_MODEL"))
	if gemini == "" {
		gemini = DefaultGeminiModel
	}
	primary := gemini
	if strings.TrimSpace(os.Getenv("AI_CHAT_PRIMARY_MODEL")) == LunaModel {
		primary = LunaModel
	}
	return AIChat{geminiModel: gemini, primaryModel: primary, geminiFallback: os.Getenv("AI_CHAT_FALLBACK") == "gemini"}
}

// GeminiModel returns the configured Gemini model independently of the primary.
func (c AIChat) GeminiModel() string { return c.geminiModel }

// PrimaryModel returns the selected runtime model.
func (c AIChat) PrimaryModel() string { return c.primaryModel }

// GeminiFallback reports whether the exact fallback selector is enabled.
func (c AIChat) GeminiFallback() bool { return c.geminiFallback }

// AIChatTraceEnabled evaluates the operational toggle at each call. Unknown
// values are disabled, as before, rather than being startup failures.
func AIChatTraceEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("AI_CHAT_TRACE_LOGS"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
