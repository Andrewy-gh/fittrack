package aichat

import (
	"fmt"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	telemetryCategoryStream   = "stream"
	telemetryCategoryRecovery = "recovery"
	telemetryCategoryLoad     = "load"
	telemetryCategoryUX       = "ux"

	telemetryOutcomeCompleted                 = "completed"
	telemetryOutcomeServerError               = "server_error"
	telemetryOutcomeTransportEndedPreTerminal = "transport_ended_pre_terminal"
	telemetryOutcomeClientAborted             = "client_aborted"
	telemetryOutcomeRecoveredCompleted        = "recovered_completed"
	telemetryOutcomeRecoveredFailed           = "recovered_failed"
	telemetryOutcomeRecoveryTimeout           = "recovery_timeout"
	telemetryOutcomeRecoveryAborted           = "recovery_aborted"
	telemetryOutcomeLoadCompleted             = "load_completed"
	telemetryOutcomeLoadFailed                = "load_failed"
	telemetryOutcomeLoadAbortedStale          = "load_aborted_stale"
	telemetryOutcomeFailureToastShown         = "failure_toast_shown"
	telemetryOutcomeFailureToastSuppressed    = "failure_toast_suppressed_due_to_successful_recovery"
	telemetryStreamStagePreStart              = "pre_start"
	telemetryStreamStagePostStart             = "post_start"
	telemetryStreamStageTerminal              = "terminal"
	telemetryStageNotApplicable               = "n/a"
	telemetryCohortBeta                       = "beta"
	telemetryCohortNonBeta                    = "non_beta"

	aiChatStreamMilestonePrepared   = "prepared"
	aiChatStreamMilestoneSSEStarted = "sse_started"
	aiChatStreamMilestoneFirstDelta = "first_delta"
	aiChatStreamMilestoneDone       = "done"

	aiChatStreamEventStarted          = "started"
	aiChatStreamEventFirstDelta       = "first_delta"
	aiChatStreamEventCompleted        = "completed"
	aiChatStreamEventAwaitingRecovery = "awaiting_recovery"
	aiChatStreamEventError            = "error"
	aiChatStreamEventStartWriteFailed = "start_write_failed"
	aiChatStreamEventDoneWriteFailed  = "done_write_failed"

	aiChatModelOperationStreamChat = "stream_chat"

	aiChatPersistenceOperationAppendChunk      = "append_stream_chunk"
	aiChatPersistenceOperationCompleteRun      = "complete_run"
	aiChatPersistenceOperationSaveWorkoutDraft = "save_workout_draft"

	aiChatMetricResultSuccess          = "success"
	aiChatMetricResultError            = "error"
	aiChatMetricResultTimeout          = "timeout"
	aiChatMetricResultClientDisconnect = "client_disconnect"
)

var (
	aiChatClientOutcomesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_chat_client_outcomes_total",
			Help: "Client-observed AI chat outcomes for stream, recovery, load, and UX flows.",
		},
		[]string{"category", "outcome", "stage", "cohort"},
	)

	aiChatStreamMilestoneDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_chat_stream_milestone_duration_seconds",
			Help:    "Elapsed time from an AI chat stream request start to key stream milestones.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"milestone"},
	)

	aiChatStreamEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ai_chat_stream_events_total",
			Help: "Server-observed AI chat stream lifecycle events.",
		},
		[]string{"event"},
	)

	aiChatModelDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_chat_model_duration_seconds",
			Help:    "Duration of AI chat model/provider work.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "result"},
	)

	aiChatPersistenceDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ai_chat_persistence_duration_seconds",
			Help:    "Duration of AI chat persistence operations.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "result"},
	)

	validTelemetryOutcomes = map[string]map[string]struct{}{
		telemetryCategoryStream: {
			telemetryOutcomeCompleted:                 {},
			telemetryOutcomeServerError:               {},
			telemetryOutcomeTransportEndedPreTerminal: {},
			telemetryOutcomeClientAborted:             {},
		},
		telemetryCategoryRecovery: {
			telemetryOutcomeRecoveredCompleted: {},
			telemetryOutcomeRecoveredFailed:    {},
			telemetryOutcomeRecoveryTimeout:    {},
			telemetryOutcomeRecoveryAborted:    {},
		},
		telemetryCategoryLoad: {
			telemetryOutcomeLoadCompleted:    {},
			telemetryOutcomeLoadFailed:       {},
			telemetryOutcomeLoadAbortedStale: {},
		},
		telemetryCategoryUX: {
			telemetryOutcomeFailureToastShown:      {},
			telemetryOutcomeFailureToastSuppressed: {},
		},
	}

	validStreamStages = map[string]struct{}{
		telemetryStreamStagePreStart:  {},
		telemetryStreamStagePostStart: {},
		telemetryStreamStageTerminal:  {},
	}
)

type ClientTelemetryEvent struct {
	Category string `json:"category"`
	Outcome  string `json:"outcome"`
	Stage    string `json:"stage,omitempty"`
}

func normalizeClientTelemetryEvent(event ClientTelemetryEvent) ClientTelemetryEvent {
	event.Category = strings.TrimSpace(event.Category)
	event.Outcome = strings.TrimSpace(event.Outcome)
	event.Stage = strings.TrimSpace(event.Stage)
	return event
}

func validateClientTelemetryEvent(event ClientTelemetryEvent) error {
	event = normalizeClientTelemetryEvent(event)

	outcomes, ok := validTelemetryOutcomes[event.Category]
	if !ok {
		return fmt.Errorf("unsupported telemetry category %q", event.Category)
	}

	if _, ok := outcomes[event.Outcome]; !ok {
		return fmt.Errorf("unsupported telemetry outcome %q for category %q", event.Outcome, event.Category)
	}

	if event.Category == telemetryCategoryStream {
		if _, ok := validStreamStages[event.Stage]; !ok {
			return fmt.Errorf("unsupported stream telemetry stage %q", event.Stage)
		}
		return nil
	}

	if event.Stage != "" {
		return fmt.Errorf("telemetry stage is only supported for stream outcomes")
	}

	return nil
}

func recordClientTelemetry(hasFeatureAccess bool, event ClientTelemetryEvent) {
	event = normalizeClientTelemetryEvent(event)

	stage := event.Stage
	if stage == "" {
		stage = telemetryStageNotApplicable
	}

	aiChatClientOutcomesTotal.WithLabelValues(
		event.Category,
		event.Outcome,
		stage,
		resolveTelemetryCohort(hasFeatureAccess),
	).Inc()
}

func recordAIChatStreamMilestone(milestone string, startedAt time.Time) {
	if startedAt.IsZero() {
		return
	}

	aiChatStreamMilestoneDuration.WithLabelValues(milestone).Observe(time.Since(startedAt).Seconds())
}

func recordAIChatStreamEvent(event string) {
	aiChatStreamEventsTotal.WithLabelValues(event).Inc()
}

func recordAIChatModelDuration(operation string, startedAt time.Time, result string) {
	if startedAt.IsZero() {
		return
	}

	aiChatModelDuration.WithLabelValues(operation, result).Observe(time.Since(startedAt).Seconds())
}

func recordAIChatPersistenceDuration(operation string, startedAt time.Time, result string) {
	if startedAt.IsZero() {
		return
	}

	aiChatPersistenceDuration.WithLabelValues(operation, result).Observe(time.Since(startedAt).Seconds())
}

func resolveTelemetryCohort(hasFeatureAccess bool) string {
	if hasFeatureAccess {
		return telemetryCohortBeta
	}
	return telemetryCohortNonBeta
}
