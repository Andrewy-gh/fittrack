package aichat

import "context"

type trainingProfileSourceContextKey struct{}

type trainingProfileSource struct {
	ConversationID int32
	MessageID      int32
}

func contextWithTrainingProfileSource(ctx context.Context, conversationID int32, messageID int32) context.Context {
	if conversationID <= 0 || messageID <= 0 {
		return ctx
	}
	return context.WithValue(ctx, trainingProfileSourceContextKey{}, trainingProfileSource{
		ConversationID: conversationID,
		MessageID:      messageID,
	})
}

func trainingProfileSourceFromContext(ctx context.Context) (trainingProfileSource, bool) {
	source, ok := ctx.Value(trainingProfileSourceContextKey{}).(trainingProfileSource)
	if !ok || source.ConversationID <= 0 || source.MessageID <= 0 {
		return trainingProfileSource{}, false
	}
	return source, true
}
