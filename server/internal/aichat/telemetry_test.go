package aichat

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
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

func TestRecordClientTelemetryNormalizesLabels(t *testing.T) {
	aiChatClientOutcomesTotal.Reset()

	recordClientTelemetry(true, ClientTelemetryEvent{
		Category: " stream ",
		Outcome:  " transport_ended_pre_terminal ",
		Stage:    " pre_start ",
	})

	if got := testutil.ToFloat64(aiChatClientOutcomesTotal.WithLabelValues(
		telemetryCategoryStream,
		telemetryOutcomeTransportEndedPreTerminal,
		telemetryStreamStagePreStart,
		telemetryCohortBeta,
	)); got != 1 {
		t.Fatalf("expected canonical telemetry labels to be recorded once, got %v", got)
	}
}
