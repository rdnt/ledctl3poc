package event

import (
	"github.com/google/uuid"

	"ledctl3/pkg/event"
)

type SetSourceIdleEvent struct {
	Event
	InputIds []uuid.UUID `json:"inputIds"`
}

func (e SetSourceIdleEvent) Type() event.Type {
	return SetSourceIdle
}
