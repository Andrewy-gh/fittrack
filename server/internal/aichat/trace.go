package aichat

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

const aiChatTraceLogsEnvVar = "AI_CHAT_TRACE_LOGS"

func aiChatTraceEnabled() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv(aiChatTraceLogsEnvVar)))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func logAIChatTrace(logger *slog.Logger, stage string, attrs ...any) {
	if !aiChatTraceEnabled() {
		return
	}
	if logger == nil {
		logger = slog.Default()
	}

	logAttrs := make([]any, 0, len(attrs)+2)
	logAttrs = append(logAttrs, "stage", stage)
	logAttrs = append(logAttrs, attrs...)
	logger.Info("ai chat stream trace", logAttrs...)
}

func logAIChatTraceContext(ctx context.Context, stage string, attrs ...any) {
	if !aiChatTraceEnabled() {
		return
	}

	logAttrs := make([]any, 0, len(attrs)+2)
	logAttrs = append(logAttrs, "stage", stage)
	logAttrs = append(logAttrs, attrs...)
	slog.InfoContext(ctx, "ai chat stream trace", logAttrs...)
}
