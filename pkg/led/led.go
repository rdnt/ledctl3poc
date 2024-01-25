package led

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/samber/lo"

	"ledctl3/internal/device"
	"ledctl3/internal/device/common"
	"ledctl3/pkg/uuid"
	"ledctl3/pkg/ws281x"
)

func New() device.Driver {
	return &dev{}
}

type outputConfig struct {
	Id     uuid.UUID `json:"id"`
	Offset int       `json:"offset"`
	Count  int       `json:"count"`
}

type config struct {
	Outputs []outputConfig `json:"outputs"`
}

type dev struct {
	id        uuid.UUID
	cfg       config
	reg       common.IORegistry
	store     common.StateHolder
	engine    *ws281x.Engine
	renderMux sync.Mutex
	rendering bool
}

func (d *dev) SetId(id uuid.UUID) {
	d.id = id
}

func (d *dev) Id() uuid.UUID {
	return d.id
}

func (d *dev) SetConfig(b []byte) error {
	err := d.store.SetConfig(b)
	if err != nil {
		return err
	}

	err = d.applyConfig(b)
	if err != nil {
		return err
	}

	return nil
}

func (d *dev) Config() ([]byte, error) {
	b, err := json.Marshal(d.cfg)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func defaultCfg() config {
	return config{
		Outputs: []outputConfig{},
	}
}

func (d *dev) Start(id uuid.UUID, reg common.IORegistry, store common.StateHolder) error {
	d.id = id
	d.reg = reg
	d.store = store

	b, err := d.store.GetConfig()
	if errors.Is(err, os.ErrNotExist) {
		cfg := defaultCfg()
		b, err = json.Marshal(cfg)
		if err != nil {
			return err
		}

		err := d.store.SetConfig(b)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	err = d.applyConfig(b)
	if err != nil {
		return err
	}

	return nil
}

func (d *dev) applyConfig(b []byte) error {
	var cfg config
	err := json.Unmarshal(b, &cfg)
	if err != nil {
		return err
	}

	if d.engine != nil {
		d.engine.Stop()
		d.engine.Fini()
		d.engine = nil
	}

	removed, added := lo.Difference(d.cfg.Outputs, cfg.Outputs)
	for _, o := range removed {
		fmt.Println("Removing output", o)
		d.reg.RemoveOutput(o.Id)
	}

	for _, o := range added {
		fmt.Println("Adding output", o)
		out := newOutput(o.Id, o.Count, o.Offset, d)
		d.reg.AddOutput(out)
	}

	total := 0
	for _, o := range cfg.Outputs {
		total += o.Count
	}

	fmt.Println("start engine with", total, "leds")
	engine, err := ws281x.Init(18, total, 255, "grb")
	if err != nil {
		return err
	}

	d.engine = engine

	return nil
}

func (d *dev) Render() error {
	d.renderMux.Lock()
	if d.rendering {
		d.renderMux.Unlock()
		return nil
	}
	d.rendering = true
	d.renderMux.Unlock()

	err := d.engine.Render()
	if err != nil {
		return err
	}

	d.renderMux.Lock()
	d.rendering = false
	d.renderMux.Unlock()

	//fmt.Print(".")
	return nil
}
