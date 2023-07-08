package event

import (
	"github.com/google/uuid"

	"ledctl3/pkg/event"
)

type SetSinkActiveEvent struct {
	Event
	SessionId uuid.UUID   `json:"sessionId"`
	OutputIds []uuid.UUID `json:"outputIds"`
}

func (e SetSinkActiveEvent) Type() event.Type {
	return SetSinkActive
}
