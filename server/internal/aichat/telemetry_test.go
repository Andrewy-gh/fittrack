package aichat

import (
	"testing"
	"time"

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

func TestRecordAIChatStreamEvent(t *testing.T) {
	aiChatStreamEventsTotal.Reset()

	recordAIChatStreamEvent(aiChatStreamEventCompleted)

	if got := testutil.ToFloat64(aiChatStreamEventsTotal.WithLabelValues(aiChatStreamEventCompleted)); got != 1 {
		t.Fatalf("expected completed stream event to be recorded once, got %v", got)
	}
}

func TestRecordAIChatDurationMetrics(t *testing.T) {
	aiChatStreamMilestoneDuration.Reset()
	aiChatModelDuration.Reset()
	aiChatPersistenceDuration.Reset()
	startedAt := time.Now().Add(-time.Millisecond)

	recordAIChatStreamMilestone(aiChatStreamMilestoneFirstDelta, startedAt)
	recordAIChatModelDuration(aiChatModelOperationStreamChat, startedAt, aiChatMetricResultSuccess)
	recordAIChatPersistenceDuration(aiChatPersistenceOperationCompleteRun, startedAt, aiChatMetricResultSuccess)

	if got := testutil.CollectAndCount(aiChatStreamMilestoneDuration); got != 1 {
		t.Fatalf("expected one stream milestone metric, got %v", got)
	}
	if got := testutil.CollectAndCount(aiChatModelDuration); got != 1 {
		t.Fatalf("expected one model duration metric, got %v", got)
	}
	if got := testutil.CollectAndCount(aiChatPersistenceDuration); got != 1 {
		t.Fatalf("expected one persistence duration metric, got %v", got)
	}
}
