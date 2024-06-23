package common

import (
	"github.com/google/uuid"

	"ledctl3/node/types"
)

type Input interface {
	Id() uuid.UUID
	SourceId() uuid.UUID
	//Init(cfg types.InputConfig) error
	Events() <-chan types.UpdateEvent
	//Close() error
	Schema() []byte
	LinkSchema() []byte
	//AssistedSetup() map[string]any
}

type Output interface {
	Id() uuid.UUID
	DriverId() uuid.UUID
	Leds() int
	//Init() error
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
