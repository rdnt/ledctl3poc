package event

import "ledctl3/pkg/uuid"

type SetInputConfigEvent struct {
	Event
	InputId uuid.UUID      `json:"inputId"`
	Config  map[string]any `json:"config"`
}

func (e SetInputConfigEvent) Type() Type {
	return SetInputConfig
}
