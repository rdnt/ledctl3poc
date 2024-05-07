package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Input struct {
	Id         uuid.UUID       `json:"id"`
	NodeId     uuid.UUID       `json:"node_id"`
	SourceId   uuid.UUID       `json:"source_id"`
	Name       string          `json:"name"`
	Connected  bool            `json:"connected"`
	Schema     json.RawMessage `json:"schema"`
	Config     json.RawMessage `json:"config"`
	LinkSchema json.RawMessage `json:"link_schema"`
}

func NewInput(id, nodeId, sourceId uuid.UUID) *Input {
	return &Input{
		Id:       id,
		NodeId:   nodeId,
		SourceId: sourceId,
	}
}
