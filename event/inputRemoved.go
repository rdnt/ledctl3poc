package event

import "ledctl3/pkg/uuid"

type InputRemovedEvent struct {
	Event
	Id uuid.UUID `json:"id"`
}

func (e InputRemovedEvent) Type() Type {
	return InputRemoved
}
