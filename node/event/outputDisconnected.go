package event

import "github.com/google/uuid"

type OutputDisconnected struct {
	Id     uuid.UUID
	SinkId uuid.UUID
}

func (e OutputDisconnected) Type() string {
	return TypeOutputDisconnected
}
