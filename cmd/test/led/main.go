package main

import (
	"time"

	"ledctl3/pkg/ws281x"
)

func main() {
	engine, err := ws281x.Init(18, 8, 255, "grb")
	if err != nil {
		panic(err)
	}

	for {
		var j int
		t := time.NewTicker(100 * time.Millisecond)

		for {
			select {
			case <-t.C:
				if j%10 == 0 {
					for i := range 8 {
						err = engine.SetLedColor(i, 0x1D, 0x1D, 0x1D, 0x00)
						if err != nil {
							panic(err)
						}
					}

					err = engine.Render()
					if err != nil {
						panic(err)
					}
				}

				if j%10 == 1 {
					for i := range 8 {
						err = engine.SetLedColor(i, 0x00, 0x00, 0x00, 0x00)
						if err != nil {
							panic(err)
						}
					}

					err = engine.Render()
					if err != nil {
						panic(err)
					}
				}
				j++
			}

		}
	}
}
