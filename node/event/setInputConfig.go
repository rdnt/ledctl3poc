package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ConfigureInput struct {
	SourceId uuid.UUID
	InputId  uuid.UUID
	Config   json.RawMessage
}

func (ConfigureInput) Type() string {
	return TypeConfigureInput
}
