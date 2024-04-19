package event

import "github.com/google/uuid"

type OutputConnected struct {
	Id       uuid.UUID
	DriverId uuid.UUID
	Leds     int
	Schema   []byte
	Config   []byte
}

func (e OutputConnected) Type() string {
	return TypeOutputConnected
}
