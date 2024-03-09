package req

import "ledctl3/pkg/uuid"

type SetSourceConfig struct {
	SourceId uuid.UUID
	Config   []byte
}

func (SetSourceConfig) Type() string {
	return TypeSetSourceConfig
}

type SetSinkConfig struct {
	SinkId uuid.UUID
	Config []byte
}

func (SetSinkConfig) Type() string {
	return TypeSetSinkConfig
}
