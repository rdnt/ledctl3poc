package event

import "ledctl3/pkg/uuid"

type SetSourceConfig struct {
	SourceId uuid.UUID
	Config   []byte
}

type SetSinkConfig struct {
	SinkId uuid.UUID
	Config []byte
}

type Response struct {
	Success bool
	Error   *string
}
