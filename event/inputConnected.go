package event

import "ledctl3/pkg/uuid"

type InputConnected struct {
	Id     uuid.UUID
	Schema map[string]any
	Config map[string]any
}
