package debug

import (
	"image/color"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"ledctl3/source"
)

type DebugInput struct {
	id     uuid.UUID
	events chan source.UpdateEvent
	pixs   map[uuid.UUID][]color.Color
}

func New() *DebugInput {
	return &DebugInput{
		id:     uuid.New(),
		events: make(chan source.UpdateEvent),
		pixs:   make(map[uuid.UUID][]color.Color),
	}
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
					pix := make([]color.Color, output.Leds)

					for i := 0; i < output.Leds; i++ {
						pix[i] = color.RGBA{R: 0, G: 0, B: 0}
					}

					pix[rand.Intn(output.Leds)] = color.RGBA{R: 255, G: 255, B: 255}

					i.pixs[output.Id] = pix

					outputs = append(outputs, source.UpdateOutput{
						Id:  output.Id,
						Pix: pix,
					})
				}

				i.events <- source.UpdateEvent{
					Outputs: outputs,
					SinkId:  sinkCfg.Id,
					Latency: 1000 * time.Millisecond,
				}

				time.Sleep(1000 * time.Millisecond)
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

func (i *DebugInput) Pixs() map[uuid.UUID][]color.Color {
	return i.pixs
}
