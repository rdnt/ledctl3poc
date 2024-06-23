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
	"ledctl3/node/types"
	"ledctl3/pkg/screencapture/dxgi"
	types2 "ledctl3/pkg/screencapture/types"
)

func init() {
	scr, err := New("dxgi")
	if err != nil {
		panic(err)
	}

	node.RegisterSource("screen", scr)
}

func New(typ string) (node.Source, error) {
	repo, err := newDisplayRepo(typ)
	if err != nil {
		return nil, err
	}

	c := &Capturer{
		repo:   repo,
		inputs: make(map[uuid.UUID]*input),
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

type Capturer struct {
	id    uuid.UUID
	reg   common.InputRegistry
	store common.StateHolder
	repo  types2.DisplayRepository

	inputs map[uuid.UUID]*input

	mux    sync.Mutex
	cancel context.CancelFunc
	wg     sync.WaitGroup

	//cfg  SourceSchemaJson
	//bcfg []byte
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
	c.cancel()
	c.wg.Wait()
	c.cancel = nil
}

func (c *Capturer) Id() uuid.UUID {
	return c.id
}

func (c *Capturer) Configure(cfg types.SourceConfig) error {
	c.mux.Lock()
	defer c.mux.Unlock()

	//var restart bool
	//if !bytes.Equal(c.bcfg, cfg.Config) {
	//err := c.applyConfig(cfg.Config)
	//if err != nil {
	//	return err
	//}
	//restart = true
	//}

	//c.stop()

	for _, cfg := range cfg.Inputs {
		in, ok := c.inputs[cfg.InputId]
		if !ok {
			fmt.Println("unknown input", cfg.InputId)
			continue
		}

		fmt.Println("Configuring input", in.id, len(cfg.Outputs))
		err := in.configure(cfg)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println("configured", in.id)
	}

	//c.start()

	//if restart {
	//	fmt.Println("RESTARTING ALL")
	//	c.captureCancel()
	//	c.captureWg.Wait()
	//}

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

	c.start()

	return nil
}

func (c *Capturer) start() {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	c.wg.Add(1)
	go c.mainLoop(ctx)
}

func (c *Capturer) mainLoop(ctx context.Context) {
	defer c.wg.Done()

	for {
		err := c.run2(ctx)
		if err != nil {
			fmt.Println(err)
		}
		time.Sleep(30 * time.Millisecond)
	}
}

func (c *Capturer) run2(ctx context.Context) error {
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

	fmt.Println("restarting run with displays", c.inputs)

	//ctx, cancel := context.WithCancel(context.Background())
	//c.captureCancel = cancel
	//
	//for _, in := range c.inputs {
	//	c.captureInput(ctx, in)
	//}
	//
	//fmt.Println("waiting for inputs")
	//c.captureWg.Wait()
	//fmt.Println("inputs stopped")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var captureWg sync.WaitGroup

	for _, in := range c.inputs {
		captureWg.Add(1)

		go func() {
			defer captureWg.Done()
			defer cancel()

			in.run2(ctx)
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
