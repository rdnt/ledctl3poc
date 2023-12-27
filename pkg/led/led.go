package led

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/samber/lo"

	"ledctl3/internal/device"
	"ledctl3/internal/device/common"
	"ledctl3/pkg/uuid"
	"ledctl3/pkg/ws281x"
)

func init() {
	d := &Driver{}
	device.Register(d)
}

type outputConfig struct {
	Id     uuid.UUID `json:"id"`
	Offset int       `json:"offset"`
	Count  int       `json:"count"`
}

type config struct {
	Outputs []outputConfig `json:"outputs"`
}

type Driver struct {
	cfg       config
	reg       common.IORegistry
	engine    *ws281x.Engine
	renderMux sync.Mutex
	rendering bool
}

func (d *Driver) SetConfig(b []byte) error {
	err := d.applyConfig(b)
	if err != nil {
		return err
	}

	err = os.WriteFile("./device-led.json", b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) Schema() ([]byte, error) {
	return nil, nil
}

func (d *Driver) Config() ([]byte, error) {
	b, err := json.Marshal(d.cfg)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (d *Driver) Start(reg common.IORegistry) error {
	d.reg = reg

	b, err := os.ReadFile("./device-led.json")
	if err != nil {
		return err
	}

	err = d.applyConfig(b)
	if err != nil {
		return err
	}

	return nil
}

func (d *Driver) applyConfig(b []byte) error {
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

func (d *Driver) Render() error {
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
