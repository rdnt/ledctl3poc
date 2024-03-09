package event

import "ledctl3/pkg/uuid"

type SetSourceConfigCommand struct {
	SourceId uuid.UUID
	Config   []byte
}

func (SetSourceConfigCommand) Type() string {
	return TypeSetSourceConfigCommand
}

type SetSinkConfigCommand struct {
	SinkId uuid.UUID
	Config []byte
}

func (SetSinkConfigCommand) Type() string {
	return TypeSetSinkConfigCommand
}

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
