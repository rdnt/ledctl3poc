package debug

import (
	"image/color"
	"time"

	"github.com/google/uuid"

	"ledctl3/source"
)

type DebugInput struct {
	id     uuid.UUID
	events chan source.UpdateEvent
}

func New() *DebugInput {
	return &DebugInput{id: uuid.New(), events: make(chan source.UpdateEvent)}
}

func (i *DebugInput) Id() uuid.UUID {
	return i.id
}

func (i *DebugInput) Start(cfg source.Config) error {
	for _, sinkCfg := range cfg.Sinks {

		sinkCfg := sinkCfg
		go func() {
			for {
				outputs := make([]source.UpdateOutput, 0)
				for _, output := range sinkCfg.Outputs {
					pix := make([]color.Color, output.Leds*4)

					outputs = append(outputs, source.UpdateOutput{
						Id:  output.Id,
						Pix: pix,
					})
				}

				i.events <- source.UpdateEvent{
					Outputs: outputs,
					SinkId:  sinkCfg.Id,
					Latency: 100 * time.Millisecond,
				}

				time.Sleep(100 * time.Millisecond)
			}
		}()
	}

	return nil
}

func (i *DebugInput) Events() chan source.UpdateEvent {
	return i.events
}

func (i *DebugInput) Stop() error {
	return nil
}
