package event

import "ledctl3/pkg/uuid"

type CapabilitiesEvent struct {
	Event
	Id      uuid.UUID                 `json:"id"`
	Inputs  []CapabilitiesEventInput  `json:"inputs"`
	Outputs []CapabilitiesEventOutput `json:"outputs"`
}

type CapabilitiesEventInput struct {
	Id           uuid.UUID      `json:"id"`
	Type         InputType      `json:"type"`
	ConfigSchema map[string]any `json:"configSchema"`
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
