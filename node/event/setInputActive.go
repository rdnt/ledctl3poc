package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type SetInputActive struct {
	Id      uuid.UUID
	Outputs []SetInputActiveOutput
}

type SetInputActiveOutput struct {
	OutputId uuid.UUID
	SinkId   uuid.UUID
	Leds     int
	Config   json.RawMessage
}

func (e SetInputActive) Type() string {
	return TypeSetInputActive
}
