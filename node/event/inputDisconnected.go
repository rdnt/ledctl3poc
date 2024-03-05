package event

import "ledctl3/pkg/uuid"

type InputDisconnected struct {
	Id uuid.UUID
}

func (e InputDisconnected) Type() string {
	return TypeInputDisconnected
}
