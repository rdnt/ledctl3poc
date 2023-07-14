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

	//pixs   map[uuid.UUID][]color.Color
}

func (i *DebugInput) Schema() map[string]any {
	return nil
}

func (i *DebugInput) ApplyConfig(b []byte) error {
	return nil
}

func New() *DebugInput {
	i := &DebugInput{
		id:     uuid.New(),
		events: make(chan source.UpdateEvent),
		//pixs:   make(map[uuid.UUID][]color.Color),
	}

	//go func() {
	//	for {
	//		for _, pix := range i.pixs {
	//			out := ""
	//			for _, c := range pix {
	//				r, g, b, _ := c.RGBA()
	//				out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprint(" ")
	//			}
	//			fmt.Println(out)
	//		}
	//		time.Sleep(500 * time.Millisecond)
	//	}
	//}()

	return i
}

// TODO: cursed variable name T_T
func (i *DebugInput) Id() uuid.UUID {
	return i.id
}

func (i *DebugInput) Start(cfg source.SinkConfig) error {
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

					//i.pixs[output.Id] = pix

					outputs = append(outputs, source.UpdateOutput{
						Id:  output.Id,
						Pix: pix,
					})
				}

				i.events <- source.UpdateEvent{
					Outputs: outputs,
					SinkId:  sinkCfg.Id,
					Latency: 500 * time.Millisecond,
				}

				time.Sleep(16 * time.Millisecond)
				//fmt.Println("---------------------------------------------")

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

//func (i *DebugInput) Pixs() map[uuid.UUID][]color.Color {
//	return i.pixs
//}
