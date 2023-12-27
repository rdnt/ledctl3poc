package led

import (
	"fmt"
	"image/color"

	gcolor "github.com/gookit/color"

	"ledctl3/pkg/uuid"
)

type output struct {
	id        uuid.UUID
	leds      int
	rendering bool
	driver    *Driver
	offset    int
}

func newOutput(id uuid.UUID, leds int, offset int, d *Driver) *output {
	i := &output{
		id:     id,
		leds:   leds,
		offset: offset,
		driver: d,
	}

	return i
}

func (o *output) Id() uuid.UUID {
	return o.id
}

func (o *output) Leds() int {
	return o.leds
}

func (o *output) Render(pix []color.Color) {
	out := ""
	for i, c := range pix {
		r, g, b, _ := c.RGBA()
		err := o.driver.engine.SetLedColor(i+o.offset, uint8(r>>8), uint8(g>>8), uint8(b>>8), 255)
		if err != nil {
			fmt.Println("error setting led color:", err)
			continue
		}
		out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprint(" ")
	}
	fmt.Println(out)

	err := o.driver.Render()
	if err != nil {
		fmt.Println("error rendering:", err)
		return
	}
}
