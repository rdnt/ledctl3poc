package event

import "ledctl3/pkg/uuid"

type AssistedSetupConfig struct {
	SourceId uuid.UUID
	InputId  uuid.UUID
	Config   any
}
