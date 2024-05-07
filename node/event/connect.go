package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Capabilities struct {
	Sources []CapabilitiesSource
	Sinks   []CapabilitiesSink
}

func (e Capabilities) Type() string {
	return TypeCapabilities
}

type CapabilitiesSource struct {
	Id     uuid.UUID
	Schema json.RawMessage
}

type CapabilitiesSink struct {
	Id     uuid.UUID
	Schema json.RawMessage
}
