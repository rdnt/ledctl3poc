package event

import "ledctl3/pkg/uuid"

type AssistedSetupEvent struct {
	Event
	InputId uuid.UUID `json:"inputId"`
}

func (e AssistedSetupEvent) Type() Type {
	return AssistedSetup
}
