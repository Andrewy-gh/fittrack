package aichat

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Andrewy-gh/fittrack/server/internal/user"
	"github.com/inngest/inngestgo"
)

const (
	inngestAppID             = "fittrack-api"
	inngestRecoveryEventName = "fittrack/ai-chat.run-recovery.requested"
	inngestRecoveryFunction  = "ai-chat-run-recovery"
)

type inngestRecoveryEventData struct {
	ConversationID int32  `json:"conversation_id"`
	RunID          int32  `json:"run_id"`
	UserID         string `json:"user_id"`
	Reason         string `json:"reason,omitempty"`
}

type recoveryRunner interface {
	RecoverStreamingRun(ctx context.Context, request RunRecoveryRequest) error
}

type InngestRecovery struct {
	client inngestgo.Client
}

func NewInngestRecovery(logger *slog.Logger, runner recoveryRunner) (*InngestRecovery, error) {
	client, err := inngestgo.NewClient(inngestgo.ClientOpts{
		AppID:  inngestAppID,
		Logger: logger,
	})
	if err != nil {
		return nil, fmt.Errorf("create inngest client: %w", err)
	}

	_, err = inngestgo.CreateFunction(
		client,
		inngestgo.FunctionOpts{
			ID:          inngestRecoveryFunction,
			Name:        "AI chat run recovery",
			Idempotency: inngestgo.StrPtr("event.data.run_id"),
			Retries:     inngestgo.IntPtr(3),
		},
		inngestgo.EventTrigger(inngestRecoveryEventName, nil),
		func(ctx context.Context, input inngestgo.Input[inngestRecoveryEventData]) (any, error) {
			req := RunRecoveryRequest{
				ConversationID: input.Event.Data.ConversationID,
				RunID:          input.Event.Data.RunID,
				UserID:         input.Event.Data.UserID,
				Reason:         input.Event.Data.Reason,
			}

			if err := runner.RecoverStreamingRun(user.WithContext(ctx, req.UserID), req); err != nil {
				return nil, err
			}

			return RecoverMessageResponse{
				ConversationID: req.ConversationID,
				RunID:          req.RunID,
				Status:         recoverStatusQueued,
			}, nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("register inngest recovery function: %w", err)
	}

	return &InngestRecovery{client: client}, nil
}

func (r *InngestRecovery) EnqueueRunRecovery(ctx context.Context, request RunRecoveryRequest) error {
	eventID := fmt.Sprintf("ai-chat-run-recovery-%d", request.RunID)
	_, err := r.client.Send(ctx, inngestgo.GenericEvent[inngestRecoveryEventData]{
		ID:   &eventID,
		Name: inngestRecoveryEventName,
		Data: inngestRecoveryEventData{
			ConversationID: request.ConversationID,
			RunID:          request.RunID,
			UserID:         request.UserID,
			Reason:         request.Reason,
		},
	})
	if err != nil {
		return fmt.Errorf("send inngest recovery event: %w", err)
	}

	return nil
}

func (r *InngestRecovery) Handler() http.Handler {
	path := "/inngest"
	return r.client.ServeWithOpts(inngestgo.ServeOpts{Path: &path})
}
