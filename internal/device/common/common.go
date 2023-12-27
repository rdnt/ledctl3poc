package common

import (
	"image/color"

	"ledctl3/internal/device/types"
	"ledctl3/pkg/uuid"
)

type Input interface {
	Id() uuid.UUID
	Start(cfg types.InputConfig) error
	Events() <-chan types.UpdateEvent
	Stop() error
	Schema() map[string]any
	AssistedSetup() map[string]any
}

type Output interface {
	Id() uuid.UUID
	Render([]color.Color)
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
