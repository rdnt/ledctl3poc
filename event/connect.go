package event

import (
	"encoding/json"

	"ledctl3/pkg/uuid"
)

type Connect struct {
	Id      uuid.UUID
	Sources []ConnectSource
	Sinks   []ConnectSink
}

type ConnectSource struct {
	Id     uuid.UUID
	Config json.RawMessage
	Schema json.RawMessage
}

type ConnectSink struct {
	Id     uuid.UUID
	Config json.RawMessage
	Schema json.RawMessage
}
