package aichat

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	generationStatusQueued      = "queued"
	generationStatusGenerating  = "generating"
	generationStatusCompleted   = "completed"
	generationStatusFailed      = "failed"
	generationStatusInterrupted = "interrupted"

	generationHeartbeatInterval = 5 * time.Second
	generationLeaseDuration     = 30 * time.Second
	chatGenerationTimeout       = 75 * time.Second
	maxGenerationAttempts       = 2

	runOwnerKindAPI     = "api"
	runOwnerKindInngest = "inngest"
)

type runOwner struct {
	value string
}

func newAPIRunOwner() runOwner {
	host, err := os.Hostname()
	if err != nil || strings.TrimSpace(host) == "" {
		host = "unknown"
	}
	return newRunOwner(runOwnerKindAPI, fmt.Sprintf("%s-%d", host, os.Getpid()))
}

func newInngestRunOwner(runID int32) runOwner {
	return newRunOwner(runOwnerKindInngest, fmt.Sprintf("run-%d-%s", runID, uuid.NewString()))
}

func newRunOwner(kind string, id string) runOwner {
	kind = strings.TrimSpace(kind)
	id = strings.TrimSpace(id)
	if kind == "" {
		kind = "unknown"
	}
	if id == "" {
		id = "unknown"
	}
	return runOwner{value: kind + ":" + id}
}

func (o runOwner) Value() string {
	return o.value
}

func (o runOwner) LeaseExpiresAt(now time.Time) time.Time {
	return now.UTC().Add(generationLeaseDuration)
}
