package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

//type SourceConfig struct {
//	Config json.RawMessage
//	//Inputs []InputConfig
//}
//
//type InputConfig struct {
//	Config json.RawMessage
//	//Outputs []OutputConfig
//}

type OutputConfig struct {
	OutputId uuid.UUID
	NodeId   uuid.UUID
	SinkId   uuid.UUID
	Leds     int
	Config   json.RawMessage
}
