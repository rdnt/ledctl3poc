package event

import "ledctl3/pkg/uuid"

type Capabilities struct {
	Inputs  []CapabilitiesInput
	Outputs []CapabilitiesOutput
}

type CapabilitiesInput struct {
	Id     uuid.UUID
	Type   InputType
	Schema map[string]any
	Config map[string]any
}

type CapabilitiesOutput struct {
	Id   uuid.UUID
	Leds int
}
