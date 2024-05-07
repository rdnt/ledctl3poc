package types

import (
	"encoding/json"

	"github.com/google/uuid"
)

type Link struct {
	Id          uuid.UUID       `json:"id"`
	ProfileId   uuid.UUID       `json:"profile_id"`
	InputId     uuid.UUID       `json:"input_id"`
	OutputId    uuid.UUID       `json:"output_id"`
	InputConfig json.RawMessage `json:"input_config"`
}
