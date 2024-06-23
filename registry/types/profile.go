package types

import (
	"github.com/google/uuid"
)

type Profile struct {
	Id      uuid.UUID   `json:"id"`
	Name    string      `json:"name"`
	Enabled bool        `json:"-"`
	LinkIds []uuid.UUID `json:"link_ids"`
}
