package debug

import (
	"github.com/google/uuid"

	"ledctl3/sink"
)

type DebugOutput struct {
	id     uuid.UUID
	events chan sink.UpdateEvent

	//pixs   map[uuid.UUID][]color.Color
}

func New() *DebugOutput {
	i := &DebugOutput{
		id:     uuid.New(),
		events: make(chan sink.UpdateEvent),
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
func (o *DebugOutput) Id() uuid.UUID {
	return o.id
}

func (o *DebugOutput) Start() error {

	return nil
}

func (o *DebugOutput) Events() chan sink.UpdateEvent {
	return o.events
}

func (o *DebugOutput) Stop() error {
	return nil
}

//func (i *DebugOutput) Pixs() map[uuid.UUID][]color.Color {
//	return i.pixs
//}
