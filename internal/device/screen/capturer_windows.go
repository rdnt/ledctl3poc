package screen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/samber/lo"

	"ledctl3/internal/device"
	"ledctl3/internal/device/common"
	"ledctl3/internal/device/types"
	"ledctl3/pkg/screencapture/dxgi"
	types2 "ledctl3/pkg/screencapture/types"
	"ledctl3/pkg/uuid"
)

func New(typ string) (device.Driver, error) {
	repo, err := newDisplayRepo(typ)
	if err != nil {
		return nil, err
	}

	c := &Capturer{
		repo:      repo,
		inputs:    make(map[uuid.UUID]*Input),
		captureWg: new(sync.WaitGroup),
	}

	return c, nil
}

func newDisplayRepo(typ string) (types2.DisplayRepository, error) {
	switch typ {
	case "dxgi":
		return dxgi.New()
	default:
		return nil, errors.New("invalid capturer type")
	}
}

type InputRegistry interface {
	AddInput(i common.Input)
	RemoveInput(id uuid.UUID)
}

type State struct {
	Associations map[string][]uuid.UUID `json:"associations"`
}

type config struct {
	CapturerType string `json:"capturerType"`
}

type Capturer struct {
	id            uuid.UUID
	reg           common.InputRegistry
	store         common.StateHolder
	repo          types2.DisplayRepository
	inputs        map[uuid.UUID]*Input
	capturing     bool
	captureCtx    context.Context
	captureCancel context.CancelFunc
	captureWg     *sync.WaitGroup
	cfg           any
}

func (c *Capturer) Id() uuid.UUID {
	return c.id
}

func (c *Capturer) SetConfig(cfg []byte) error {
	err := c.applyConfig(b)
	if err != nil {
		return err
	}

	err = os.WriteFile("./device-screen.json", b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *Capturer) applyConfig(b []byte) error {
	var cfg config
	err := json.Unmarshal(b, &cfg)
	if err != nil {
		return err
	}

	repo, err := newDisplayRepo(cfg.CapturerType)
	if err != nil {
		return err
	}

	c.repo = repo

	return nil
}

func (c *Capturer) Schema() ([]byte, error) {
	return nil, nil
}

func (c *Capturer) Config() ([]byte, error) {
	b, err := json.Marshal(c.cfg)
	if err != nil {
		return nil, err
	}

	return b, nil
}
func (c *Capturer) Start(id uuid.UUID, reg common.IORegistry, store common.StateHolder) error {
	c.id = id
	c.reg = reg
	c.store = store

	for {
		err := c.init()
		if err != nil {
			fmt.Println(err)
			continue
		}

		break
	}

	for _, in := range c.inputs {
		fmt.Printf("%#v\n", in.capturer)
	}

	//fmt.Println("Capturer initialized")

	go func() {
		for {
			err := c.run()
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(1 * time.Second)
		}
	}()

	return nil
}

func displayAssociationId(ds []types2.Display) string {
	var ids []string
	for _, d := range ds {
		id := fmt.Sprintf("%d,%d,%d,%d,%d,%d", d.Id(), d.Width(), d.Height(), d.X(), d.Y(), d.Orientation())
		ids = append(ids, id)
	}

	return strings.Join(ids, "|")
}

func (c *Capturer) setState(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}

	return c.store.SetState(b)
}

func (c *Capturer) getState(v any) error {
	b, err := c.store.GetState()
	if err != nil {
		return err
	}

	err = json.Unmarshal(b, &v)
	if err != nil {
		return err
	}

	return nil
}

func (c *Capturer) init() error {
	displays, err := c.repo.All()
	if err != nil {
		return err
	}

	if len(displays) == 0 {
		fmt.Println("No displays")
		return nil
	}

	s := State{}
	err = c.getState(&s)
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

		err = c.setState(s)
		if err != nil {
			return err
		}

		assoc = s.Associations[assocId]
	}

	fmt.Println("Current association:", assoc)
	fmt.Println(assoc)

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

		//fmt.Println("Added input", in.uuid, "for display", d.OutputId())

		c.inputs[assoc[i]] = in
		c.reg.AddInput(in)
	}

	return nil
}

func (c *Capturer) run() error {
	//println("=== run")

	for id, in := range c.inputs {
		//c.src.RemoveInput(outputId) // TODO: race when listing capabilities
		if in.display == nil {
			continue
		}
		_ = in.display.Close()
		//in.display = nil
		//delete(c.inputs, outputId)
		_ = id
	}

	displays, err := c.repo.All()
	if err != nil {
		return err
	}

	if len(displays) == 0 {
		fmt.Println("No displays")
		return nil
	}

	s := State{}
	err = c.getState(&s)
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

		err = c.setState(s)
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
		c.reg.AddInput(in)
	}

	c.captureCtx, c.captureCancel = context.WithCancel(context.Background())

	fmt.Println("restarting capture with displays", c.inputs)

	for _, in := range c.inputs {
		c.captureWg.Add(1)
		in := in
		go func() {
			defer c.captureWg.Done()

			frames := in.display.Capture(c.captureCtx, in.cfg.Framerate)

			for frame := range frames {
				//if !in.started {
				//	fmt.Printf(" %d-", in.display.OutputId())
				//} else {
				//	fmt.Printf(" %d ", in.display.OutputId())
				//}

				if in.started {
					go in.processFrame(frame)
				}
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
//			fmt.Print(in.display.OutputId(), " ")
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
