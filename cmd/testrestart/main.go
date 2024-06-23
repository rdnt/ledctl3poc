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

	fmt.Println("1")
	func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		frames := d.Capture(ctx, 60)

		go func() {
			defer cancel()

			for range frames {
				fmt.Println(".")
			}
		}()
	}()

	time.Sleep(5 * time.Second)

	fmt.Println("2")
	func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		frames := d.Capture(ctx, 60)

		go func() {
			defer func() {
				fmt.Println("@@@@@@@@@@@@@@@@@@@@ CLOSE 1")
			}()
			defer cancel()

			for range frames {
				fmt.Println(".")
			}
		}()
	}()

	time.Sleep(15 * time.Second)
	fmt.Println("close")
	d.Close()
}
