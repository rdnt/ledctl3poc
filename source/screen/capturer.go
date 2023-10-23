package screen

import (
	"context"
	"fmt"
	"strings"
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
	SetState(any) error
	GetState(any) error
}

type State struct {
	Associations map[string][]uuid.UUID `json:"associations"`
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
		for {
			err := c.run()
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(1 * time.Second)
		}
	}()
}

func displayAssociationId(ds []types2.Display) string {
	var ids []string
	for _, d := range ds {
		id := fmt.Sprintf("%d,%d,%d,%d,%d,%d", d.Id(), d.Width(), d.Height(), d.X(), d.Y(), d.Orientation())
		ids = append(ids, id)
	}

	return strings.Join(ids, "|")
}

func (c *Capturer) run() error {
	println("=== run")

	//for _, in := range c.inputs {
	//	//c.src.RemoveInput(uid)
	//	_ = in.display.Close()
	//	in.display = nil
	//	//delete(c.inputs, uid)
	//}

	displays, err := c.repo.All()
	if err != nil {
		return err
	}

	s := State{}
	err = c.src.GetState(&s)
	if err != nil {
		return err
	}

	assocId := displayAssociationId(displays)

	assoc, ok := s.Associations[assocId]
	if !ok {
		if s.Associations == nil {
			s.Associations = make(map[string][]uuid.UUID)
		}

		s.Associations[assocId] = []uuid.UUID{}
		for range displays {
			s.Associations[assocId] = append(s.Associations[assocId], uuid.New())
		}

		err = c.src.SetState(s)
		if err != nil {
			return err
		}

		assoc = s.Associations[assocId]
	}

	//for _, in := range c.inputs {
	//	if !slices.Contains(assoc, in.uuid) {
	//		in.started = false
	//	}
	//}

	for i, d := range displays {
		if in, ok := c.inputs[assoc[i]]; ok {
			in.display = d
			//c.capturing = c.capturing || in.started
			continue
		}

		in := &Input{
			capturer: c,
			uuid:     assoc[i],
			events:   make(chan types.UpdateEvent),
			display:  d,
			outputs:  nil,
			cfg: types.InputConfig{
				Framerate: 1,
				Outputs:   nil,
			},
		}

		fmt.Println("Added input", in.uuid, "for display", d.Id())

		c.inputs[assoc[i]] = in
		c.src.AddInput(in)
	}

	c.captureCtx, c.captureCancel = context.WithCancel(context.Background())

	for _, in := range c.inputs {
		c.captureWg.Add(1)
		in := in
		go func() {
			defer c.captureWg.Done()

			frames := in.display.Capture(c.captureCtx, in.cfg.Framerate)

			for frame := range frames {
				if !in.started {
					fmt.Print(in.display.Id(), "- ")
				} else {
					fmt.Print(in.display.Id(), "  ")
				}

				go in.processFrame(frame)
			}

			// capture was cancelled or errored. restart with possibly new config
			c.captureCancel()
		}()
	}

	c.captureWg.Wait()

	return nil
}

//func (c *Capturer) run() {
//	println("=== run")
//	c.capturing = false
//
//	// TODO: re-query inputs from repo
//	// TODO: try to re-pair old uuids with new inputs, based on resolution, numeric IDs etc.
//
//	for _, in := range c.inputs {
//		if !in.started {
//			continue
//		}
//
//		c.captureInput(in)
//	}
//}

//func (c *Capturer) startInput(in *Input, cfg types.InputConfig) error {
//	//println("=== startInput")
//	//
//	//in.cfg = cfg
//	//
//	//ctx, cancel := context.WithCancel(context.Background())
//	//in.cancel = cancel
//	//in.ctx = ctx
//	//// TODO: ramp up framerate
//	//in.cancel()
//	////c.captureInput(in)
//	////c.captureCancel()
//	//
//	//return nil
//}

//func (c *Capturer) captureInput(in *Input) {
//	println("=== captureInput")
//
//	go func() {
//		defer c.captureWg.Done()
//
//		//framerate := 1
//		//if in.cfg.Framerate > 0 {
//		//	framerate = in.cfg.Framerate
//		//}
//		frames := in.display.Capture(c.captureCtx, 1)
//
//		for frame := range frames {
//			//fmt.Print(in.uuid, " ")
//			fmt.Print(in.display.Id(), " ")
//			//fmt.Println(in.display)
//
//			go in.processFrame(frame)
//		}
//
//		c.captureCancel()
//	}()
//}

func (c *Capturer) Inputs() ([]*Input, error) {
	return lo.Values(c.inputs), nil
}
