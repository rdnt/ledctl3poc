package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Source struct {
	Id        uuid.UUID       `json:"id"`
	NodeId    uuid.UUID       `json:"node_id"`
	Name      string          `json:"name"`
	Connected bool            `json:"connected"`
	Schema    json.RawMessage `json:"schema"`
	Config    json.RawMessage `json:"config"`
	InputIds  []uuid.UUID     `json:"input_ids"`
}

func NewSource(id, nodeId uuid.UUID) *Source {
	return &Source{
		Id:     id,
		NodeId: nodeId,
	}
}
