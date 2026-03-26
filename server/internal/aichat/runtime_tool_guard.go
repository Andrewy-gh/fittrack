package aichat

import (
	"context"
	"sync"
)

type toolGuardContextKey struct{}

type toolCallGuard struct {
	mu                 sync.Mutex
	featureSnapshot    *featureSnapshot
	featureSnapshotErr error
}

func withToolGuard(ctx context.Context) context.Context {
	return context.WithValue(ctx, toolGuardContextKey{}, &toolCallGuard{})
}

func toolGuardFromContext(ctx context.Context) *toolCallGuard {
	if ctx == nil {
		return nil
	}
	guard, _ := ctx.Value(toolGuardContextKey{}).(*toolCallGuard)
	return guard
}

func (g *toolCallGuard) listActiveFeatures(ctx context.Context, featureAccess featureAccessReader) (*featureSnapshot, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.featureSnapshot != nil || g.featureSnapshotErr != nil {
		return cloneFeatureSnapshot(g.featureSnapshot), g.featureSnapshotErr
	}

	snapshot, err := listActiveFeaturesSnapshot(ctx, featureAccess)
	g.featureSnapshot = cloneFeatureSnapshot(snapshot)
	g.featureSnapshotErr = err

	return cloneFeatureSnapshot(snapshot), err
}

func cloneFeatureSnapshot(snapshot *featureSnapshot) *featureSnapshot {
	if snapshot == nil {
		return nil
	}

	keys := make([]string, len(snapshot.FeatureKeys))
	copy(keys, snapshot.FeatureKeys)

	return &featureSnapshot{FeatureKeys: keys}
}
