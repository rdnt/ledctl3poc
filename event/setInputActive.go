package event

import (
	"ledctl3/pkg/uuid"
)

type SetInputActive struct {
	Id      uuid.UUID
	Outputs []SetInputActiveOutput
}

type SetInputActiveOutput struct {
	OutputId uuid.UUID
	DeviceId uuid.UUID
	SinkId   uuid.UUID
	Leds     int
	Config   map[string]any
}
