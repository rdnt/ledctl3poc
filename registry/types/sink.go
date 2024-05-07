package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Sink struct {
	Id        uuid.UUID       `json:"id"`
	NodeId    uuid.UUID       `json:"node_id"`
	Name      string          `json:"name"`
	Connected bool            `json:"connected"`
	Schema    json.RawMessage `json:"schema"`
	Config    json.RawMessage `json:"config"`
	OutputIds []uuid.UUID     `json:"output_ids"`
}

func NewSink(id, nodeId uuid.UUID) *Sink {
	return &Sink{
		Id:     id,
		NodeId: nodeId,
	}
}
