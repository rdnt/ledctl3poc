package event

import (
	"github.com/google/uuid"
)

type SetSourceIdleEvent struct {
	Event
	InputIds []uuid.UUID `json:"inputIds"`
}

func (e SetSourceIdleEvent) Type() Type {
	return SetSourceIdle
}
