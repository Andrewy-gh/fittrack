package aichat

import "testing"

func TestAIChatTraceEnabled(t *testing.T) {
	t.Run("disabled by default", func(t *testing.T) {
		t.Setenv(aiChatTraceLogsEnvVar, "")

		if aiChatTraceEnabled() {
			t.Fatal("expected AI chat trace logs to be disabled by default")
		}
	})

	t.Run("accepts common truthy values", func(t *testing.T) {
		for _, value := range []string{"1", "true", "TRUE", "yes", "on"} {
			t.Run(value, func(t *testing.T) {
				t.Setenv(aiChatTraceLogsEnvVar, value)

				if !aiChatTraceEnabled() {
					t.Fatalf("expected %s=%q to enable AI chat trace logs", aiChatTraceLogsEnvVar, value)
				}
			})
		}
	})

	t.Run("rejects non-truthy values", func(t *testing.T) {
		for _, value := range []string{"0", "false", "off", "no"} {
			t.Run(value, func(t *testing.T) {
				t.Setenv(aiChatTraceLogsEnvVar, value)

				if aiChatTraceEnabled() {
					t.Fatalf("expected %s=%q to disable AI chat trace logs", aiChatTraceLogsEnvVar, value)
				}
			})
		}
	})
}
