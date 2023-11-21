package event

import "ledctl3/pkg/uuid"

type InputAdded struct {
	Id     uuid.UUID
	Type   InputType
	Schema map[string]any
	Config map[string]any
}
