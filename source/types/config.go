package types

import "ledctl3/pkg/uuid"

type InputConfig struct {
	Framerate int
	Outputs   []OutputConfig
}

type OutputConfig struct {
	Id     uuid.UUID
	SinkId uuid.UUID
	Config map[string]any
	Leds   int
}
