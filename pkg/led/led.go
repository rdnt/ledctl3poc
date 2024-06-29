package led

import (
	"context"
	"encoding/json"
	"fmt"
	"image/color"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"ledctl3/node"
	"ledctl3/node/common"
	"ledctl3/pkg/ws281x"
)

func init() {
	node.RegisterSink("led", New())
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
	id     uuid.UUID
	cfg    config
	reg    common.IORegistry
	store  common.StateHolder
	engine *ws281x.Engine

	cancelRender context.CancelFunc

	wg sync.WaitGroup
	//renderMux sync.Mutex
	//rendering bool
	//queue     bool
	outputs map[uuid.UUID]*output

	leds int
	pix  []color.NRGBA

	queued bool

	pixMux    sync.Mutex
	rendering bool

	invalidateCh chan bool
}

func New() node.Sink {
	return &dev{
		outputs:      map[uuid.UUID]*output{},
		invalidateCh: make(chan bool, 1),
	}
}

func (d *dev) SetId(id uuid.UUID) {
	d.id = id
}

func (d *dev) Id() uuid.UUID {
	return d.id
}

func (d *dev) Configure(b []byte) error {
	err := d.applyConfig(b)
	if err != nil {
		return err
	}

	return nil
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

	return nil
}

func (d *dev) startRenderQueue() {
	ctx, cancel := context.WithCancel(context.Background())
	d.cancelRender = cancel

	d.wg.Add(1)
	go func() {
		defer d.wg.Done()

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				d.invalidate()
			case <-d.invalidateCh:
				d.render()
			}
		}
	}()
}

func (d *dev) stopRenderQueue() {
	if d.cancelRender == nil {
		return
	}

	d.cancelRender()
	d.wg.Wait()
}

func (d *dev) invalidate() {
	select {
	case d.invalidateCh <- true:
	default:
	}
}

func (d *dev) Stop() error {
	return nil
}

func (d *dev) applyConfig(b []byte) error {
	var cfg config
	err := json.Unmarshal(b, &cfg)
	if err != nil {
		return err
	}

	if d.engine != nil {
		d.stopRenderQueue()
		d.engine.Fini()
		d.engine = nil
	}

	removed, added := lo.Difference(d.cfg.Outputs, cfg.Outputs)

	for _, o := range removed {
		fmt.Println("Removing output", o)
		d.reg.RemoveOutput(o.Id)
		d.leds -= d.outputs[o.Id].leds
		delete(d.outputs, o.Id)
	}

	for _, o := range added {
		//fmt.Println("Adding output", o)
		out := newOutput(o.Id, o.Count, o.Offset, d)
		d.reg.AddOutput(out)
		d.leds += out.leds
		d.outputs[o.Id] = out
	}

	d.pix = nil

	d.cfg = cfg

	total := 0
	for _, o := range cfg.Outputs {
		total += o.Count
	}

	//fmt.Println("start engine with", total, "leds")
	//fmt.Println("-")
	engine, err := ws281x.Init(18, total, 180, "grb") // TODO: make configurable
	if err != nil {
		return err
	}

	d.engine = engine

	d.startRenderQueue()

	return nil
}

func (d *dev) Render(pixs map[uuid.UUID][]color.NRGBA) error {
	outPixs := make(map[int][]color.NRGBA)

	for outputId, pix := range pixs {
		out, ok := d.outputs[outputId]
		if !ok {
			continue
		}

		if out.Leds() != len(pix) {
			continue
		}

		outPixs[out.offset] = pix
	}

	d.pixMux.Lock()

	if d.pix == nil {
		d.pix = make([]color.NRGBA, d.leds)
	}

	for offset, pix := range outPixs {
		for i := range len(pix) {
			d.pix[offset+i] = pix[i]
		}
	}

	d.pixMux.Unlock()

	d.invalidate()

	return nil
}

func (d *dev) render() {
	d.pixMux.Lock()

	if d.pix == nil {
		d.pixMux.Unlock()
		return
	}

	for i, c := range d.pix {
		r, g, b, _ := c.RGBA()
		err := d.engine.SetLedColor(i, uint8(r>>8), uint8(g>>8), uint8(b>>8), 255)
		if err != nil {
			fmt.Println("error setting led color:", err)
			continue
		}
		//out += gcolor.RGB(uint8(r>>8), uint8(g>>8), uint8(b>>8), true).Sprint(" ")
	}

	d.pixMux.Unlock()

	err := d.engine.Render()
	if err != nil {
		fmt.Println(err)
	}
}
