package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ConfigureSource struct {
	SourceId     uuid.UUID
	SourceConfig json.RawMessage
	Inputs       []ConfigureSourceInput
}

type ConfigureSourceInput struct {
	InputId     uuid.UUID
	InputConfig json.RawMessage
	Outputs     []ConfigureSourceOutput
}

type ConfigureSourceOutput struct {
	NodeId   uuid.UUID
	SinkId   uuid.UUID
	OutputId uuid.UUID
	Leds     int
	Config   json.RawMessage
}

func (e ConfigureSource) Type() string {
	return TypeConfigureSource
}
