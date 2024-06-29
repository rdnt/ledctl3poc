package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ConfigureOutput struct {
	SinkId   uuid.UUID
	OutputId uuid.UUID
	Config   json.RawMessage
}

func (ConfigureOutput) Type() string {
	return TypeConfigureOutput
}
