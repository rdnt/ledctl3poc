package event

import "github.com/google/uuid"

type SetInputConfigEvent struct {
	Event
	InputId uuid.UUID      `json:"inputId"`
	Config  map[string]any `json:"config"`
}

func (e SetInputConfigEvent) Type() Type {
	return SetInputConfig
}
