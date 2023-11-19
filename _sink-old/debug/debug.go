package debug

import (
	"ledctl3/pkg/uuid"

	"ledctl3/_sink-old"
)

type DebugOutput struct {
	id     uuid.UUID
	events chan _sink_old.UpdateEvent
	leds   int

	//pixs   map[uuid.UUID][]color.Color
}

func New(id uuid.UUID, leds int) *DebugOutput {
	i := &DebugOutput{
		id:     id,
		leds:   leds,
		events: make(chan _sink_old.UpdateEvent),
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

func (o *DebugOutput) Id() uuid.UUID {
	return o.id
}

func (o *DebugOutput) Start() error {
	return nil
}

func (o *DebugOutput) Leds() int {
	return o.leds
}

func (o *DebugOutput) Events() chan _sink_old.UpdateEvent {
	return o.events
}

func (o *DebugOutput) Stop() error {
	return nil
}

//func (i *DebugOutput) Pixs() map[uuid.UUID][]color.Color {
//	return i.pixs
//}
