package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Output struct {
	Id        uuid.UUID       `json:"id"`
	NodeId    uuid.UUID       `json:"node_id"`
	SinkId    uuid.UUID       `json:"sink_id"`
	Name      string          `json:"name"`
	Leds      int             `json:"leds"`
	Connected bool            `json:"connected"`
	Schema    json.RawMessage `json:"schema"`
	Config    json.RawMessage `json:"config"`
}

func NewOutput(id, nodeId, sinkId uuid.UUID, leds int) *Output {
	return &Output{
		Id:     id,
		NodeId: nodeId,
		SinkId: sinkId,
		Leds:   leds,
	}
}
