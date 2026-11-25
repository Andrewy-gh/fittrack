package aichat

import (
	"context"
	"log/slog"

	"github.com/Andrewy-gh/fittrack/server/internal/config"
)

func logAIChatTrace(logger *slog.Logger, stage string, attrs ...any) {
	if !config.AIChatTraceEnabled() {
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
	if !config.AIChatTraceEnabled() {
		return
	}

	logAttrs := make([]any, 0, len(attrs)+2)
	logAttrs = append(logAttrs, "stage", stage)
	logAttrs = append(logAttrs, attrs...)
	slog.InfoContext(ctx, "ai chat stream trace", logAttrs...)
}
