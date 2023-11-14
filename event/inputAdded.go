package event

import "ledctl3/pkg/uuid"

type InputAddedEvent struct {
	Event
	Id uuid.UUID `json:"id"`
}

func (e InputAddedEvent) Type() Type {
	return InputAdded
}
