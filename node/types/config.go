package types

import "github.com/google/uuid"

type InputConfig struct {
	Framerate int
	Outputs   []OutputConfig
}

type OutputConfig struct {
	Id     uuid.UUID
	NodeId uuid.UUID
	SinkId uuid.UUID
	//Config map[string]any
	Config OutputConfigConfig
	Leds   int
}

type OutputConfigConfig struct {
	Width   int
	Height  int
	Left    int
	Top     int
	Reverse bool
}
