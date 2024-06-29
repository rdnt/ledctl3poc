package screen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"ledctl3/node"
	"ledctl3/node/common"
	"ledctl3/pkg/screencapture/dxgi"
	types2 "ledctl3/pkg/screencapture/types"
)

func init() {
	node.RegisterSource("screen", New())
}

func New() node.Source {
	c := &Capturer{
		inputs: make(map[uuid.UUID]*input),
	}

	return c
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

type Capturer struct {
	id    uuid.UUID
	reg   common.InputRegistry
	store common.StateHolder
	repo  types2.DisplayRepository

	inputs map[uuid.UUID]*input

	mux    sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup

	cfg SourceConfig
}

type SourceConfig SourceSchemaJson

func (s SourceConfig) Equal(s2 SourceConfig) bool {
	return s.CapturerType == s2.CapturerType
}

func (c *Capturer) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.cancel == nil {
		panic("already closed")
		return nil
	}

	return nil
}

func (c *Capturer) stop() {
	fmt.Println("capturer.stop")

	if c.cancel == nil {
		return
	}

	c.cancel()
	c.wg.Wait()
	c.cancel = nil
}

func (c *Capturer) Id() uuid.UUID {
	return c.id
}

func (c *Capturer) Configure(b []byte) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	var cfg SourceConfig
	err := json.Unmarshal(b, &cfg)
	if err != nil {
		return err
	}

	if c.cfg.Equal(cfg) {
		//return nil
	}

	c.stop()

	fmt.Printf("source config set %+v\n", c.cfg)

	err = c.start()
	if err != nil {
		return err
	}

	return nil
}

func (c *Capturer) Init(id uuid.UUID, reg common.IORegistry, store common.StateHolder) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	if c.cancel != nil {
		c.cancel()
		c.wg.Wait()
		c.cancel = nil
	}

	c.id = id
	c.reg = reg
	c.store = store

	c.cfg = SourceConfig{
		CapturerType: "dxgi",
	}

	err := c.start()
	if err != nil {
		return err
	}

	return nil
}

func (c *Capturer) start() error {
	fmt.Println("capturer.start")

	repo, err := newDisplayRepo(string(c.cfg.CapturerType))
	if err != nil {
		return err
	}

	c.repo = repo

	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	c.wg.Add(1)
	go c.mainLoop(ctx)

	return nil
}

func (c *Capturer) mainLoop(ctx context.Context) {
	fmt.Println("capturer.mainLoop")

	defer c.wg.Done()

	for {
		select {
		default:
			fmt.Println("capturer.mainLoop.iter")
			err := c.run(ctx)
			if err != nil {
				fmt.Println(err)
			}
			time.Sleep(30 * time.Millisecond)
			if ctx.Err() != nil {
				return
			}
		}
	}
}

func (c *Capturer) run(ctx context.Context) error {
	fmt.Println("capturer.run")

	displays, err := c.repo.All()
	if err != nil {
		fmt.Println("repo.All error", err)
		c.disconnectInputs()
		return err
	}

	if len(displays) == 0 {
		fmt.Println("No displays")
		c.disconnectInputs()
		return nil
	}

	s := State{}
	err = c.getState(&s)
	if err != nil {
		fmt.Println("getState error", err)
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

	for id, in := range c.inputs {
		// delete disconnected inputs
		if !slices.Contains(assoc, in.id) {
			c.reg.RemoveInput(in.id)
			delete(c.inputs, id)
		}
	}

	for i, d := range displays {
		// assign display to existing input
		if in, ok := c.inputs[assoc[i]]; ok {
			in.display = d
			continue
		}

		// create new input
		in := newInput(assoc[i], c.id, d)

		//in.applyConfig(InputConfig{
		//	Config: InputSchemaJson{
		//		Framerate: 1,
		//	},
		//	Outputs: nil,
		//})
		//
		fmt.Println("Added input", in.id, "for display", d.Id())

		c.inputs[assoc[i]] = in
		c.reg.AddInput(in)
	}

	//fmt.Println("restarting run with displays", c.inputs)

	//ctx, cancel := context.WithCancel(context.Background())
	//c.captureCancel = cancel
	//
	//for _, in := range c.inputs {
	//	c.captureInput(ctx, in)
	//}
	//
	//fmt.Println("waiting for inputs")
	//c.captureWg.Wait()
	fmt.Println("inputs starting")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var captureWg sync.WaitGroup

	for _, in := range c.inputs {
		captureWg.Add(1)

		go func() {
			defer captureWg.Done()
			defer cancel()

			in.run(ctx)
		}()
	}

	captureWg.Wait()

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

func (c *Capturer) disconnectInputs() {
	fmt.Println("Disconnecting inputs")

	for id, in := range c.inputs {
		c.reg.RemoveInput(in.id)
		delete(c.inputs, id)
	}
}

//func (c *Capturer) run() error {
//	println("=== run")
//
//	displays, err := c.repo.All()
//	if err != nil {
//		fmt.Println("repo.All error", err)
//		c.disconnectInputs()
//		return err
//	}
//
//	if len(displays) == 0 {
//		fmt.Println("No displays")
//		c.disconnectInputs()
//		return nil
//	}
//
//	s := State{}
//	err = c.getState(&s)
//	if err != nil {
//		return err
//	}
//
//	fmt.Println("INIT")
//
//	assocId := displayAssociationId(displays)
//
//	assoc, ok := s.Associations[assocId]
//	if !ok {
//		if s.Associations == nil {
//			s.Associations = make(map[string][]uuid.UUID)
//		}
//
//		s.Associations[assocId] = []uuid.UUID{}
//		for range displays {
//			s.Associations[assocId] = append(s.Associations[assocId], uuid.New())
//		}
//
//		err = c.setState(s)
//		if err != nil {
//			return err
//		}
//
//		assoc = s.Associations[assocId]
//	}
//
//	for id, in := range c.inputs {
//		// delete disconnected inputs
//		if !slices.Contains(assoc, in.id) {
//			c.reg.RemoveInput(in.id)
//			delete(c.inputs, id)
//		}
//	}
//
//	for i, d := range displays {
//		// assign display to existing input
//		if in, ok := c.inputs[assoc[i]]; ok {
//			in.display = d
//			continue
//		}
//
//		// create new input
//		in := newInput(assoc[i], c.id, d)
//
//		//in.applyConfig(InputConfig{
//		//	Config: InputSchemaJson{
//		//		Framerate: 1,
//		//	},
//		//	Outputs: nil,
//		//})
//		//
//		fmt.Println("Added input", in.id, "for display", d.Id())
//
//		c.inputs[assoc[i]] = in
//		c.reg.AddInput(in)
//	}
//
//	fmt.Println("restarting run with displays", c.inputs)
//
//	//ctx, cancel := context.WithCancel(context.Background())
//	//c.captureCancel = cancel
//	//
//	//for _, in := range c.inputs {
//	//	c.captureInput(ctx, in)
//	//}
//	//
//	//fmt.Println("waiting for inputs")
//	//c.captureWg.Wait()
//	//fmt.Println("inputs stopped")
//
//	for _, d := range displays {
//		d.Close()
//	}
//
//	return nil
//}

//func (c *Capturer) captureInput(ctx context.Context, in *input) {
//	c.captureWg.Add(1)
//
//	go func() {
//		defer c.captureWg.Done()
//		defer c.captureCancel()
//
//		//in.parentCtx = ctx
//		//in.capture()
//	}()
//}
