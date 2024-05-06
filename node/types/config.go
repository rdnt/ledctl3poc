package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

type LinksConfig struct {
	Outputs []OutputConfig
}

type OutputConfig struct {
	Id     uuid.UUID
	NodeId uuid.UUID
	SinkId uuid.UUID
	Config json.RawMessage
	Leds   int
}

//type LinkConfig struct {
//	Width   int
//	Height  int
//	Left    int
//	Top     int
//	Reverse bool
//}
