package dxgi

import (
	"context"
	"errors"
	"fmt"
	"image"
	"runtime"
	"time"

	"github.com/kirides/screencapture/d3d"
	"github.com/kirides/screencapture/win"

	"ledctl3/pkg/screencapture/types"
)

var errNoFrame = errors.New("no frame")

type display struct {
	index       int
	id          int
	width       int
	height      int
	x           int
	y           int
	buf         *image.NRGBA
	dev         *d3d.ID3D11Device
	devCtx      *d3d.ID3D11DeviceContext
	ddup        *d3d.OutputDuplicator
	orientation types.Orientation
}

func (d *display) Id() int {
	return d.id
}

func (d *display) Width() int {
	return d.width
}

func (d *display) Height() int {
	return d.height
}

func (d *display) X() int {
	return d.x
}

func (d *display) Y() int {
	return d.y
}

func (d *display) Resolution() string {
	return fmt.Sprintf("%dx%d", d.width, d.height)
}

func (d *display) String() string {
	return fmt.Sprintf("Display{id: %d, width: %d, height: %d, left: %d, top: %d}", d.id, d.width, d.height, d.x, d.y)
}

func (d *display) Capture(ctx context.Context, pollRate int) (chan []byte, chan error) {
	frames := make(chan []byte, 1)
	errCh := make(chan error, 1)

	//defer d.reset()

	//fmt.Println("reset now")
	//err := d.reset()
	//if err != nil {
	//	panic(err)
	//}

	fmt.Println("@@@")
	fmt.Println(d.dev)
	fmt.Println(d.devCtx)
	fmt.Println(d.ddup)

	go func() {
		defer close(frames)
		defer close(errCh)

		ticker := time.NewTicker(time.Duration(1000/pollRate) * time.Millisecond)
		defer ticker.Stop()

		fmt.Println("============ START fps", pollRate)

		for range ticker.C {
			select {
			case <-ctx.Done():
				fmt.Println(d.id, "context done")
				return
			default:
				pix, err := d.GetFrame()
				if errors.Is(err, errNoFrame) {
					//fmt.Println(d.id, "no frame")
					continue
				} else if err != nil {

					err := d.Close()
					if err != nil {
						fmt.Println(d.id, "failed to reset from capture")
					}
					errCh <- err

					return
				}

				if pix == nil {
					panic("invalid frame")
				}

				//fmt.Println(d.id, "dispatch")
				frames <- pix
			}
		}
	}()

	return frames, errCh
}

func (d *display) Orientation() types.Orientation {
	return d.orientation
}

func (d *display) GetFrame() ([]byte, error) {
	//fmt.Println(d.id, "GetFrame")
	err := d.ddup.GetImage(d.buf, 0)
	if errors.Is(err, d3d.ErrNoImageYet) {
		// don't update
		return nil, errNoFrame
	} else if err != nil {
		return nil, err
	}

	return d.buf.Pix, nil
}

func (d *display) reset() error {
	_ = d.Close()

	// Keep this thread, so windows/d3d11/dxgi can use their threadlocal caches, if any
	runtime.LockOSThread()

	// Make thread PerMonitorV2 Dpi aware if supported on OS
	// allows to let windows handle BGRA -> RGBA conversion and possibly more things
	if win.IsValidDpiAwarenessContext(win.DpiAwarenessContextPerMonitorAwareV2) {
		_, err := win.SetThreadDpiAwarenessContext(win.DpiAwarenessContextPerMonitorAwareV2)
		if err != nil {
			fmt.Printf("Could not set thread DPI awareness to PerMonitorAwareV2. %v\n", err)
		}
	}

	var err error
	d.dev, d.devCtx, err = d3d.NewD3D11Device()
	if err != nil {
		return err
	}

	d.ddup, err = d3d.NewIDXGIOutputDuplication(d.dev, d.devCtx, uint(d.id))
	if err != nil {
		d.dev.Release()
		d.dev = nil

		d.devCtx.Release()
		d.devCtx = nil
		return err
	}

	switch d.ddup.Orientation() {
	case d3d.DXGI_MODE_ROTATION_UNSPECIFIED, d3d.DXGI_MODE_ROTATION_IDENTITY:
		d.orientation = types.Landscape
	case d3d.DXGI_MODE_ROTATION_ROTATE90:
		d.orientation = types.Portrait
	case d3d.DXGI_MODE_ROTATION_ROTATE180:
		d.orientation = types.LandscapeFlipped
	case d3d.DXGI_MODE_ROTATION_ROTATE270:
		d.orientation = types.PortraitFlipped
	}

	return nil
}

func (d *display) Close() error {
	fmt.Println("CLOSE called")
	if d.dev != nil {
		d.dev.Release()
		d.dev = nil
	}

	if d.devCtx != nil {
		d.devCtx.Release()
		d.devCtx = nil
	}

	if d.ddup != nil {
		d.ddup.Release()
		d.ddup = nil
	}

	return nil
}
