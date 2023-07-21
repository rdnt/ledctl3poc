package types

import "github.com/google/uuid"

type SinkConfig struct {
	Framerate int
	Sinks     []SinkConfigSink
}

type SinkConfigSink struct {
	Id      uuid.UUID
	Outputs []SinkConfigSinkOutput
}

type SinkConfigSinkOutput struct {
	Id     uuid.UUID
	Config map[string]any
	Leds   int
}
