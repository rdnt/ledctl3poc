package types

import (
	"github.com/google/uuid"
)

type Node struct {
	Id        uuid.UUID   `json:"id"`
	Name      string      `json:"name"`
	Connected bool        `json:"connected"`
	Address   string      `json:"address"`
	SourceIds []uuid.UUID `json:"source_ids"`
	SinkIds   []uuid.UUID `json:"sink_ids"`
}

func NewNode(id uuid.UUID) *Node {
	return &Node{
		Id: id,
	}
}
