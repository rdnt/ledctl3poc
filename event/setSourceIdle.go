package event

import (
	"github.com/google/uuid"
)

type SetSourceIdleEvent struct {
	Event
	Inputs []SetSourceIdleEventInput `json:"inputs"`
}

type SetSourceIdleEventInput struct {
	InputId   uuid.UUID   `json:"inputId"`
	OutputIds []uuid.UUID `json:"outputIds"`
}

func (e SetSourceIdleEvent) Type() Type {
	return SetSourceIdle
}
