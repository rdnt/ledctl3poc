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
	NodeId   uuid.UUID
	SinkId   uuid.UUID
	Leds     int
	Config   map[string]any
}

func (e SetInputActive) Type() string {
	return TypeSetInputActive
}
