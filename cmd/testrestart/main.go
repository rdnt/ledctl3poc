package main

import (
	"context"
	"fmt"
	"time"

	"ledctl3/pkg/screencapture/dxgi"
)

func main() {
	repo, err := dxgi.New()
	if err != nil {
		panic(err)
	}

	displays, err := repo.All()
	if err != nil {
		panic(err)
	}

	d := displays[0]

	for {
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)

			frames, errCh := d.Capture(ctx, 240)

			defer cancel()

			for {
				select {
				case _ = <-errCh:
					//fmt.Println("error during capture:", err)
					return
				case <-frames:
					fmt.Print(".")
				}
			}
		}()
	}
}
