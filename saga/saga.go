package saga

import (
	"github.com/berkaroad/uuid"
	"time"
)

type BaseSagaState struct {
	EventId   uuid.UUID
	StartTime time.Time
}

type SagaStateTracker interface {
	IsTimeout() bool
	IsCompleted() bool
}
