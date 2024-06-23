package screen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"sync"
	"time"

	"github.com/bamiaux/rez"
	"github.com/google/uuid"
	"golang.org/x/image/draw"

	"ledctl3/node/types"
	types2 "ledctl3/pkg/screencapture/types"
)

type input struct {
	id      uuid.UUID
	display types2.Display
	srcId   uuid.UUID
	events  chan types.UpdateEvent

	parentCtx     context.Context
	captureWg     sync.WaitGroup
	captureCancel context.CancelFunc

	processWg     sync.WaitGroup
	processCancel context.CancelFunc

	stop chan bool

	frames chan []byte

	bcfg []byte
	cfg  InputConfig

	prepareOnce sync.Once
	thumb       *image.NRGBA
	prescaler   draw.Scaler
	resizer     rez.Converter

	procs map[uuid.UUID]*outputProcessor

	ready     bool
	lastFrame []byte
}

type InputConfig struct {
	Config  InputSchemaJson
	Outputs map[uuid.UUID]OutputConfig
}

type OutputConfig struct {
	OutputId uuid.UUID
	NodeId   uuid.UUID
	SinkId   uuid.UUID
	Leds     int
	Config   LinkSchemaJson
}

func (in *input) Id() uuid.UUID {
	return in.id
}

func (in *input) SourceId() uuid.UUID {
	return in.srcId
}

func (in *input) Events() <-chan types.UpdateEvent {
	return in.events
}

func newInput(id, srcId uuid.UUID, d types2.Display) *input {
	in := &input{
		id:      id,
		srcId:   srcId,
		display: d,
		frames:  make(chan []byte),
		procs:   make(map[uuid.UUID]*outputProcessor),
		events:  make(chan types.UpdateEvent),
		stop:    make(chan bool),
		cfg: InputConfig{
			Config: InputSchemaJson{
				Framerate: 1,
			},
			Outputs: nil,
		},
	}

	return in
}

func (in *input) configure(cfg types.InputConfig) error {
	var restartCapture bool
	if !bytes.Equal(in.bcfg, cfg.Config) {
		restartCapture = true
	}

	var cfg2 InputSchemaJson
	err := json.Unmarshal(cfg.Config, &cfg2)
	if err != nil {
		return err
	}

	inCfg := InputConfig{
		Config:  cfg2,
		Outputs: make(map[uuid.UUID]OutputConfig),
	}

	for _, out := range cfg.Outputs {
		var ocfg LinkSchemaJson
		err := json.Unmarshal(out.Config, &ocfg)
		if err != nil {
			return err
		}

		inCfg.Outputs[out.OutputId] = OutputConfig{
			OutputId: out.OutputId,
			NodeId:   out.NodeId,
			SinkId:   out.SinkId,
			Leds:     out.Leds,
			Config:   ocfg,
		}
	}

	in.stopProcess()

	if restartCapture {
		in.stopCapture()
		in.lastFrame = nil
	}

	in.cfg = inCfg
	in.bcfg = cfg.Config

	for _, out := range in.cfg.Outputs {
		in.procs[out.OutputId] = newOutputProcessor(out, in.display.Width(), in.display.Height())
	}

	in.startProcess()

	if restartCapture {
		in.startCapture()
	}

	return nil
}

//func (in *input) run(parentCtx context.Context) {
//	fmt.Println("RUN CALLED")
//	in.parentCtx = parentCtx
//
//	in.startProcess()
//	in.startCapture()
//
//	in.mainWg.Wait()
//
//	in.stopCapture()
//	in.stopProcess()
//}

func (in *input) run2(parentCtx context.Context) {
	fmt.Println("run2")

	in.parentCtx = parentCtx

	in.startProcess()
	in.startCapture()

	<-in.stop

	in.stopCapture()
	in.stopProcess()
}

func (in *input) startCapture() {
	ctx, cancel := context.WithCancel(context.Background())
	in.captureCancel = cancel

	in.captureWg.Add(1)
	go in.capture(ctx)
}

func (in *input) stopCapture() {
	in.captureCancel()
	in.captureWg.Wait()
}

func (in *input) capture(ctx context.Context) {
	defer in.captureWg.Done()

	frames, errCh := in.display.Capture(ctx, in.cfg.Config.Framerate)
	//defer in.display.Close()

	for {
		select {
		case <-in.parentCtx.Done():
			return
		case err := <-errCh:
			if err != nil {
				in.stop <- true
			}
		case frame, ok := <-frames:
			if !ok {
				return
			}

			select {
			case in.frames <- frame:
			default:
			}
		}
	}
}

func (in *input) startProcess() {
	if in.processCancel != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	in.processCancel = cancel

	if len(in.procs) > 0 {
		in.prepareOnce.Do(func() {
			in.prepareProcess()
		})
	}

	if !in.ready {
		return
	}

	in.processWg.Add(1)
	go in.process(ctx)
}

func (in *input) prepareProcess() {
	width := in.display.Width()
	height := in.display.Height()

	prescaler := draw.BiLinear.NewScaler(width, height, width/8, height/8)

	thumb := image.NewNRGBA(image.Rect(0, 0, width/8, height/8))

	resizeCfg, err := rez.PrepareConversion(thumb, image.NewRGBA(image.Rect(0, 0, width, height)))
	if err != nil {
		fmt.Println(err)
		return
	}

	resizer, err := rez.NewConverter(resizeCfg, rez.NewBilinearFilter())
	if err != nil {
		fmt.Println(err)
		return
	}

	in.thumb = thumb
	in.resizer = resizer
	in.prescaler = prescaler
	in.ready = true
}

func (in *input) stopProcess() {
	if in.processCancel == nil {
		return
	}

	in.processCancel()
	in.processWg.Wait()

	in.procs = make(map[uuid.UUID]*outputProcessor)

	in.processCancel = nil
}

func (in *input) process(ctx context.Context) {
	defer in.processWg.Done()

	if in.lastFrame != nil {
		in.processFrame(in.lastFrame)
	}

	for {
		select {
		case <-ctx.Done():
			return
		case frame, ok := <-in.frames:
			in.lastFrame = frame

			if !ok {
				panic("frames closed")
			}

			in.processFrame(frame)
		}
	}
}

func (in *input) processFrame(pix []byte) {
	if len(in.procs) == 0 {
		return
	}

	now := time.Now()

	outs := map[uuid.UUID]map[uuid.UUID][]types.UpdateEventOutput{}

	src := &image.RGBA{
		Pix:    pix,
		Stride: in.display.Width() * 4,
		Rect:   image.Rect(0, 0, in.display.Width(), in.display.Height()),
	}

	err := in.resizer.Convert(in.thumb, src)
	if err != nil {
		fmt.Println("error resizing:", err)
		return
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(in.procs))
	var outMux sync.Mutex

	for _, out := range in.procs {
		out := out
		go func() {
			defer wg.Done()

			colors := out.process(in.thumb)

			outMux.Lock()

			if outs[out.nodeId] == nil {
				outs[out.nodeId] = map[uuid.UUID][]types.UpdateEventOutput{}
			}

			outs[out.nodeId][out.sinkId] = append(outs[out.nodeId][out.sinkId], types.UpdateEventOutput{
				SinkId:   out.sinkId,
				OutputId: out.outputId,
				Pix:      colors,
			})
			outMux.Unlock()
		}()
	}

	wg.Wait()

	for nodeId, sinkOuts := range outs {

		outputs := []types.UpdateEventOutput{}
		for sinkId, outs := range sinkOuts {
			for _, out := range outs {
				outputs = append(outputs, types.UpdateEventOutput{
					SinkId:   sinkId,
					OutputId: out.OutputId,
					Pix:      out.Pix,
				})
			}
		}

		in.events <- types.UpdateEvent{
			NodeId:  nodeId,
			Outputs: outputs,
			Latency: time.Since(now),
		}
	}
}
