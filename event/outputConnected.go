package event

import "ledctl3/pkg/uuid"

type OutputConnected struct {
	Id       uuid.UUID
	DriverId uuid.UUID
	Leds     int
	Schema   []byte
	Config   []byte
}
