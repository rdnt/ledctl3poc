package debug_output

import (
	"fmt"
	"image/color"

	"ledctl3/pkg/uuid"
)

type DebugOutput struct {
	id   uuid.UUID
	leds int
}

func (o *DebugOutput) DriverId() uuid.UUID {
	//TODO implement me
	panic("implement me")
}

func New(id uuid.UUID, leds int) *DebugOutput {
	i := &DebugOutput{
		id:   id,
		leds: leds,
	}

	return i
}

func (o *DebugOutput) Id() uuid.UUID {
	return o.id
}

func (o *DebugOutput) Leds() int {
	return o.leds
}

func (o *DebugOutput) Render(pix []color.Color) {
	fmt.Print(".")
	//out := ""
	//for _, c := range pix {
	//	r, g, b, _ := c.RGBA()
	//	out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprint(" ")
	//}
	//fmt.Println(out)
}
