package event

import (
	"ledctl3/pkg/uuid"
)

type SetSinkActiveEvent struct {
	Event
	SessionId uuid.UUID   `json:"sessionId"`
	OutputIds []uuid.UUID `json:"outputIds"`
}

func (e SetSinkActiveEvent) Type() Type {
	return SetSinkActive
}
