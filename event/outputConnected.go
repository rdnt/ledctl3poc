package event

import "ledctl3/pkg/uuid"

type OutputConnected struct {
	Id     uuid.UUID
	Leds   int
	Schema map[string]any
	Config map[string]any
}
