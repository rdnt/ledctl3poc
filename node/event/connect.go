package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Capabilities struct {
	Sources []ConnectedSource
	Sinks   []ConnectedSink
}

func (e Capabilities) Type() string {
	return TypeCapabilities
}

type ConnectedSource struct {
	Id     uuid.UUID
	Config json.RawMessage
	Schema json.RawMessage
}

type ConnectedSink struct {
	Id     uuid.UUID
	Config json.RawMessage
	Schema json.RawMessage
}
