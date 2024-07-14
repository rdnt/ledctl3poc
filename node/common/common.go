package common

import (
	"github.com/google/uuid"

	"ledctl3/node/types"
)

type Input interface {
	Id() uuid.UUID
	SourceId() uuid.UUID
	//start(cfg types.InputConfig) error
	Configure(cfg []byte) error
	Events() <-chan types.UpdateEvent
	//Close() error
	Schema() []byte
	LinkSchema() []byte
	Connect(outs []types.OutputConfig) error
	//AssistedSetup() map[string]any
}

type Output interface {
	Id() uuid.UUID
	SinkId() uuid.UUID
	Leds() int
	Configure([]byte) error
	Schema() []byte
	//start() error
	//Close() error
}

type InputRegistry interface {
	AddInput(i Input)
	RemoveInput(id uuid.UUID)
}

type OutputRegistry interface {
	AddOutput(i Output)
	RemoveOutput(id uuid.UUID)
}

type IORegistry interface {
	InputRegistry
	OutputRegistry
}

type StateHolder interface {
	SetState([]byte) error
	GetState() ([]byte, error)
}
