package common

import (
	"image/color"

	"github.com/google/uuid"

	"ledctl3/node/types"
)

type Input interface {
	Id() uuid.UUID
	DriverId() uuid.UUID
	Start(cfg types.InputConfig) error
	Events() <-chan types.UpdateEvent
	Stop() error
	Schema() map[string]any
	AssistedSetup() map[string]any
}

type Output interface {
	Id() uuid.UUID
	DriverId() uuid.UUID
	Render([]color.NRGBA)
	Leds() int
	//Start() error
	//Stop() error
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
	//Id() uuid.UUID
	SetConfig([]byte) error // TODO: only store state, and have the node react to config changes instead.
	GetConfig() ([]byte, error)
	SetState([]byte) error
	GetState() ([]byte, error)
}
