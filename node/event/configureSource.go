package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ConfigureSource struct {
	SourceId     uuid.UUID
	SourceConfig json.RawMessage
	//Inputs       []ConfigureSourceInput
}

func (e ConfigureSource) Type() string {
	return TypeConfigureSource
}

type ConfigureLinks struct {
	SourceId uuid.UUID
	InputId  uuid.UUID
	Outputs  []ConfigureLinksOutput
}

type ConfigureLinksOutput struct {
	NodeId   uuid.UUID
	SinkId   uuid.UUID
	OutputId uuid.UUID
	Leds     int
	Config   json.RawMessage
}

func (e ConfigureLinks) Type() string {
	return TypeConfigureLinks
}
