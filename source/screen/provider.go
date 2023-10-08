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

type InputRegistry interface {
	AddInput(i source.Input)
	RemoveInput(id uuid.UUID)
}

type Capturer struct {
	reg           InputRegistry
	repo          types2.DisplayRepository
	inputs        map[uuid.UUID]*Input
	captureCancel context.CancelFunc
	captureWg     *sync.WaitGroup
}

func New(reg InputRegistry) (*Capturer, error) {
	dr, err := dxgi.New()
	if err != nil {
		return nil, err
	}

	c := &Capturer{
		reg:       reg,
		repo:      dr,
		inputs:    make(map[uuid.UUID]*Input),
		captureWg: &sync.WaitGroup{},
	}

	return c, nil
}

func (c *Capturer) Start() {
	go func() {
		for {
			err := c.run()
			if err != nil {
				panic(err)
			}

			time.Sleep(1 * time.Second)
		}
	}()
}

func (c *Capturer) run() error {
	for uid, in := range c.inputs {
		c.reg.RemoveInput(uid)
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
			uuid:    uid,
			events:  make(chan types.UpdateEvent),
			display: d,
			repo:    c.repo,
			outputs: nil,
		}

		c.inputs[uid] = in
		c.reg.AddInput(in)
	}

	c.startCapture()

	return nil
}

func (c *Capturer) startCapture() {
	ctx, cancel := context.WithCancel(context.Background())
	c.captureCancel = cancel

	c.captureWg.Add(len(c.inputs))
	for _, in := range c.inputs {
		go func(in *Input) {
			in.cancel = c.captureCancel

			defer c.captureWg.Done()

			frames := in.display.Capture(ctx, 60) // TODO: framerate

			for frame := range frames {
				fmt.Print(".")

				go in.processFrame(in.display, frame)
			}

			c.captureCancel()
		}(in)
	}

	c.captureWg.Wait()
}

func (c *Capturer) Inputs() ([]*Input, error) {
	return lo.Values(c.inputs), nil
}
