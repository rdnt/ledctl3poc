package screen

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/samber/lo"

	"ledctl3/pkg/screencapture/dxgi"
	types2 "ledctl3/pkg/screencapture/types"
	"ledctl3/pkg/uuid"
	"ledctl3/source"
	"ledctl3/source/types"
)

type Source interface {
	AddInput(i source.Input)
	RemoveInput(id uuid.UUID)
}

type Capturer struct {
	src    Source
	repo   types2.DisplayRepository
	inputs map[uuid.UUID]*Input
	//captureCancel context.CancelFunc
	captureWg *sync.WaitGroup
}

func New(src Source) (*Capturer, error) {
	dr, err := dxgi.New()
	if err != nil {
		return nil, err
	}

	c := &Capturer{
		src:       src,
		repo:      dr,
		inputs:    make(map[uuid.UUID]*Input),
		captureWg: &sync.WaitGroup{},
	}

	return c, nil
}

func (c *Capturer) Start() {
	go func() {
		//for {
		err := c.run()
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(1 * time.Second)
		//}
	}()
}

func (c *Capturer) run() error {
	for uid, in := range c.inputs {
		c.src.RemoveInput(uid)
		_ = in.display.Close()
		delete(c.inputs, uid)
	}

	displays, err := c.repo.All()
	if err != nil {
		return err
	}

	for _, d := range displays {
		uid := uuid.New()
		in := &Input{
			capturer: c,
			uuid:     uid,
			events:   make(chan types.UpdateEvent),
			display:  d,
			outputs:  nil,
		}

		c.inputs[uid] = in
		c.src.AddInput(in)
	}

	return nil
}

func (c *Capturer) startInput(in *Input, cfg types.InputConfig) error {
	// stop all active inputs
	for _, in2 := range c.inputs {
		if in2.cancel != nil {
			in2.cancel()
		}

		c.captureWg.Wait()
	}

	in.capturing = true

	// start all active inputs
	for _, in2 := range c.inputs {
		if !in2.capturing {
			continue
		}

		c.captureWg.Add(1)

		ctx, cancel := context.WithCancel(context.Background())
		in2.cancel = cancel

		in2 := in2
		go func() {
			defer c.captureWg.Done()

			frames := in2.display.Capture(ctx, cfg.Framerate)

			for frame := range frames {
				//fmt.Println(in.display.Resolution())
				fmt.Print(".")

				go in2.processFrame(frame)
			}

			cancel()
		}()
	}

	return nil
}

//func (c *Capturer) startCaptureOLD() {
//	ctx, cancel := context.WithCancel(context.Background())
//	c.captureCancel = cancel
//
//	c.captureWg.Add(len(c.inputs))
//	for _, in := range c.inputs {
//		go func(in *Input) {
//			defer c.captureWg.Done()
//
//			frames := in.display.Capture(ctx, 60) // TODO: framerate
//
//			for frame := range frames {
//				fmt.Print(".")
//
//				go in.processFrame(in.display, frame)
//			}
//
//			c.captureCancel()
//		}(in)
//	}
//
//	c.captureWg.Wait()
//}

func (c *Capturer) Inputs() ([]*Input, error) {
	return lo.Values(c.inputs), nil
}
