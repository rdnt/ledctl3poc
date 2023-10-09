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
	src           Source
	repo          types2.DisplayRepository
	inputs        map[uuid.UUID]*Input
	capturing     bool
	captureCtx    context.Context
	captureCancel context.CancelFunc
	captureWg     *sync.WaitGroup
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
		err := c.run()
		if err != nil {
			fmt.Println(err)
		}

		time.Sleep(1 * time.Second)
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

func (c *Capturer) reset() {
	println("=== reset")
	c.capturing = false

	// TODO: re-query inputs from repo
	// TODO: try to re-pair old uuids with new inputs, based on resolution, numeric IDs etc.

	for _, in := range c.inputs {
		if !in.started {
			continue
		}

		c.captureInput(in)
	}
}

func (c *Capturer) startInput(in *Input, cfg types.InputConfig) error {
	println("=== startInput")

	in.started = true
	in.cfg = cfg

	c.captureInput(in)

	return nil
}

func (c *Capturer) captureInput(in *Input) {
	println("=== captureInput")

	c.captureWg.Add(1)

	if !c.capturing {
		c.capturing = true
		c.captureCtx, c.captureCancel = context.WithCancel(context.Background())

		go func() {
			c.captureWg.Wait()
			c.reset()
		}()
	}

	go func() {
		defer c.captureWg.Done()

		frames := in.display.Capture(c.captureCtx, in.cfg.Framerate)

		for frame := range frames {
			//fmt.Println(in.display.Resolution())
			fmt.Println(in.display)

			go in.processFrame(frame)
		}

		c.captureCancel()
	}()
}

func (c *Capturer) Inputs() ([]*Input, error) {
	return lo.Values(c.inputs), nil
}
