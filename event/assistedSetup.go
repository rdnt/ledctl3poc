package event

import "github.com/google/uuid"

type AssistedSetupEvent struct {
	Event
	InputId uuid.UUID `json:"inputId"`
}

func (e AssistedSetupEvent) Type() Type {
	return AssistedSetup
}
