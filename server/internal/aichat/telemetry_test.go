package aichat

import (
	"testing"
)

func TestValidateClientTelemetryEvent(t *testing.T) {
	t.Run("accepts non-stream events without stage", func(t *testing.T) {
		err := validateClientTelemetryEvent(ClientTelemetryEvent{
			Category: telemetryCategoryRecovery,
			Outcome:  telemetryOutcomeRecoveredCompleted,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("rejects stream events without stage", func(t *testing.T) {
		err := validateClientTelemetryEvent(ClientTelemetryEvent{
			Category: telemetryCategoryStream,
			Outcome:  telemetryOutcomeCompleted,
		})
		if err == nil {
			t.Fatal("expected an error for missing stream stage")
		}
	})

	t.Run("rejects non-stream stage values", func(t *testing.T) {
		err := validateClientTelemetryEvent(ClientTelemetryEvent{
			Category: telemetryCategoryUX,
			Outcome:  telemetryOutcomeFailureToastShown,
			Stage:    telemetryStreamStageTerminal,
		})
		if err == nil {
			t.Fatal("expected an error for non-stream stage usage")
		}
	})
}
