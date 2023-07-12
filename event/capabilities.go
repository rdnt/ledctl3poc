package event

import "github.com/google/uuid"

type CapabilitiesEvent struct {
	Event
	Id      uuid.UUID                 `json:"id"`
	Inputs  []CapabilitiesEventInput  `json:"inputs"`
	Outputs []CapabilitiesEventOutput `json:"outputs"`
}

type CapabilitiesEventInput struct {
	Id   uuid.UUID `json:"id"`
	Type InputType `json:"type"`
}

type CapabilitiesEventOutput struct {
	Id   uuid.UUID `json:"id"`
	Leds int       `json:"leds"`
}

type InputType string

const (
	InputTypeDefault       InputType = "default"
	InputTypeMonitor       InputType = "monitor"
	InputTypeAudioCapturer InputType = "audioCapturer"
)

func (e CapabilitiesEvent) Type() Type {
	return Capabilities
}