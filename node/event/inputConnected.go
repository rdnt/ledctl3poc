package event

import "github.com/google/uuid"

type InputConnected struct {
	Id       uuid.UUID
	DriverId uuid.UUID
	Schema   []byte
	Config   []byte
}

func (e InputConnected) Type() string {
	return TypeInputConnected
}
