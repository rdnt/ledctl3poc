package event

import "github.com/google/uuid"

type InputDisconnected struct {
	SourceId uuid.UUID
	Id       uuid.UUID
}

func (e InputDisconnected) Type() string {
	return TypeInputDisconnected
}
