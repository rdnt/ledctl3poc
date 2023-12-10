package debug_output_ws

import (
	"fmt"
	"image/color"

	gcolor "github.com/gookit/color"

	"ledctl3/pkg/uuid"
	"ledctl3/pkg/ws281x"
)

type DebugOutput struct {
	id        uuid.UUID
	leds      int
	ws        *ws281x.Engine
	rendering bool
}

func New(id uuid.UUID, leds int, ws *ws281x.Engine) *DebugOutput {
	i := &DebugOutput{
		id:   id,
		leds: leds,
		ws:   ws,
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
	if o.rendering {
		fmt.Print("!")
	}
	o.rendering = true

	out := ""
	for i, c := range pix {
		r, g, b, _ := c.RGBA()
		err := o.ws.SetLedColor(i, uint8(r>>8), uint8(g>>8), uint8(b>>8), 255)
		if err != nil {
			fmt.Println("error setting led color:", err)
			continue
		}
		out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprint(" ")
	}
	fmt.Println(out)

	err := o.ws.Render()
	if err != nil {
		fmt.Println("error rendering:", err)
		return
	}

	o.rendering = false
	//fmt.Print(".")
}
