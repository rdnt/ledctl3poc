package main

import (
	"fmt"
	"time"

	"ledctl3/pkg/screencapture/dxgi"
)

func main() {
	loopFrames()
}

func loopDisplays() {
	dr, err := dxgi.New()
	if err != nil {
		panic(err)
	}

	//now := time.Now()
	for {
		ds, err := dr.All()
		if err != nil {
			panic(err)
		}

		//fmt.Println(time.Since(now))
		//now = time.Now()
		for _, d := range ds {
			d.Close()
		}

		time.Sleep(1 * time.Second)
	}
}

func loopFrames() {
	dr, err := dxgi.New()
	if err != nil {
		panic(err)
	}

	now := time.Now()
	for {
		ds, err := dr.All()
		if err != nil {
			panic(err)
		}

		for _, d := range ds {
			fr := d.Capture()
			for range fr {
				fmt.Println(time.Since(now))
				now = time.Now()
			}
		}
	}
}
