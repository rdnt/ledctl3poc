package event

import (
	"encoding/json"

	"github.com/google/uuid"
)

type SetSourceConfigCommand struct {
	SourceId uuid.UUID
	Config   json.RawMessage
}

func (SetSourceConfigCommand) Type() string {
	return TypeSetSourceConfigCommand
}

type SetSinkConfigCommand struct {
	SinkId uuid.UUID
	Config json.RawMessage
}

func (SetSinkConfigCommand) Type() string {
	return TypeSetSinkConfigCommand
}

type SetSourceConfig struct {
	SourceId uuid.UUID
	Config   json.RawMessage
}

func (SetSourceConfig) Type() string {
	return TypeSetSourceConfig
}

type SetSinkConfig struct {
	SinkId uuid.UUID
	Config json.RawMessage
}

func (SetSinkConfig) Type() string {
	return TypeSetSinkConfig
}
