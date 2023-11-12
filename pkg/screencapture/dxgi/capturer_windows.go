package dxgi

import (
	"fmt"
	"image"

	"ledctl3/pkg/screencapture/types"
)

var scaleFactor = 8

type DxgiCapturer struct {
	displays []*display
}

func (c *DxgiCapturer) All() ([]types.Display, error) {
	fmt.Println("dxgi.All()")
	var ds []types.Display

	i := 0
	for {
		d := &display{
			id: i,
		}

		err := d.reset()
		if err != nil {
			fmt.Printf("failed to reset display %d: %v\n", i, err)
			break
		}

		bounds := d.ddup.Bounds()
		d.width = bounds.Dx()
		d.height = bounds.Dy()
		d.x = bounds.Min.X
		d.y = bounds.Min.Y

		d.buf = image.NewNRGBA(bounds)

		ds = append(ds, d)

		fmt.Printf("display after reset:\n%#v\n", d)

		i++
	}

	return ds, nil
}

func New() (*DxgiCapturer, error) {
	return &DxgiCapturer{}, nil
}
