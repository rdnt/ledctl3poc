package event

import "ledctl3/pkg/uuid"

type InputConnected struct {
	Id       uuid.UUID
	DriverId uuid.UUID
	Schema   []byte
	Config   []byte
}

func (e InputConnected) Type() string {
	return TypeInputConnected
}
