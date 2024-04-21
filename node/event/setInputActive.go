package event

import "github.com/google/uuid"

type SetInputActive struct {
	Id      uuid.UUID
	Outputs []SetInputActiveOutput
}

type SetInputActiveOutput struct {
	OutputId uuid.UUID
	SinkId   uuid.UUID
	Leds     int
	Config   map[string]any
}

func (e SetInputActive) Type() string {
	return TypeSetInputActive
}
