package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type ConfigureSink struct {
	SinkId     uuid.UUID
	SinkConfig json.RawMessage
}

func (e ConfigureSink) Type() string {
	return TypeConfigureSink
}
