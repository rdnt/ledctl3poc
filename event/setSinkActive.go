package event

import (
	"github.com/google/uuid"
)

type SetSinkActiveEvent struct {
	Event
	SessionId uuid.UUID   `json:"sessionId"`
	OutputIds []uuid.UUID `json:"outputIds"`
}

func (e SetSinkActiveEvent) Type() Type {
	return SetSinkActive
}
