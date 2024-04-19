package event

import "github.com/google/uuid"

type InputDisconnected struct {
	Id uuid.UUID
}

func (e InputDisconnected) Type() string {
	return TypeInputDisconnected
}
