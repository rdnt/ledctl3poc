package event

import (
	"encoding/json"

	"ledctl3/pkg/uuid"
)

type NodeConnected struct {
	Id      uuid.UUID
	Sources []ConnectedSource
	Sinks   []ConnectedSink
}

func (e NodeConnected) Type() string {
	return TypeNodeConnected
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
