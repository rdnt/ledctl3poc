package event

import "ledctl3/pkg/uuid"

type AssistedSetupConfigEvent struct {
	Event
	SourceId uuid.UUID      `json:"sourceId"`
	InputId  uuid.UUID      `json:"inputId"`
	Config   map[string]any `json:"config"`
}

func (e AssistedSetupConfigEvent) Type() Type {
	return AssistedSetupConfig
}
