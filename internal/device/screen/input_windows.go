package screen

import (
	"fmt"
	"image"
	"image/color"
	"sync"
	"time"

	"golang.org/x/image/draw"

	"ledctl3/pkg/uuid"

	"github.com/bamiaux/rez"

	"ledctl3/internal/device/types"
	types2 "ledctl3/pkg/screencapture/types"
)

type Input struct {
	mux      sync.Mutex
	capturer *Capturer

	uuid   uuid.UUID
	events chan types.UpdateEvent

	display   types2.Display
	outputs   map[uuid.UUID]outputCaptureConfig
	started   bool
	cfg       types.InputConfig
	prescaler draw.Scaler
	resizer   rez.Converter
	resized   *image.RGBA
}

func (in *Input) Events() <-chan types.UpdateEvent {
	return in.events
}

type outputCaptureConfig struct {
	outputId  uuid.UUID
	sinkId    uuid.UUID
	leds      int
	reverse   bool
	scaler    draw.Scaler
	subWidth  int
	subHeight int
	subLeft   int
	subTop    int
	dest      *image.RGBA
}

func (in *Input) AssistedSetup() map[string]any {
	cfg := map[string]any{
		"width":     in.display.Width(),
		"height":    in.display.Height(),
		"left":      in.display.X(),
		"top":       in.display.Y(),
		"framerate": 60,
		"reverse":   false,
	}

	return cfg
}

func (in *Input) Id() uuid.UUID {
	return in.uuid
}

func (in *Input) DriverId() uuid.UUID {
	return in.capturer.id
}

func (in *Input) Start(cfg types.InputConfig) error {
	in.mux.Lock()
	defer in.mux.Unlock()

	if fmt.Sprintf("%+v", cfg) == fmt.Sprintf("%+v", in.cfg) {
		fmt.Println("config unchanged")
		return nil
	}
	// reconfigure input and restart capture
	in.started = true

	in.cfg = cfg

	in.outputs = make(map[uuid.UUID]outputCaptureConfig)

	width := in.display.Width()
	height := in.display.Width()

	in.prescaler = draw.BiLinear.NewScaler(width, height, width/8, height/8)

	in.resized = image.NewRGBA(image.Rect(0, 0, in.display.Width()/8, in.display.Height()/8))

	resizeCfg, err := rez.PrepareConversion(in.resized, image.NewRGBA(image.Rect(0, 0, in.display.Width(), in.display.Height())))
	if err != nil {
		return err
	}

	converter, err := rez.NewConverter(resizeCfg, rez.NewBilinearFilter())
	if err != nil {
		return err
	}

	in.resizer = converter

	for _, out := range cfg.Outputs {
		reverse := out.Config.Reverse

		rect := image.Rect(out.Config.Left/8, out.Config.Top/8, (out.Config.Left+out.Config.Width)/8, (out.Config.Top+out.Config.Height)/8)

		var dst *image.RGBA

		if rect.Dx() > rect.Dy() {
			// horizontal
			dst = image.NewRGBA(image.Rect(0, 0, out.Leds, 1))
		} else {
			// vertical
			dst = image.NewRGBA(image.Rect(0, 0, 1, out.Leds))
		}

		in.outputs[out.Id] = outputCaptureConfig{
			outputId:  out.Id,
			sinkId:    out.SinkId,
			leds:      out.Leds,
			reverse:   reverse,
			scaler:    draw.BiLinear.NewScaler(width/8, height/8, out.Config.Width, out.Config.Height),
			subWidth:  out.Config.Width,
			subHeight: out.Config.Height,
			subLeft:   out.Config.Left,
			subTop:    out.Config.Top,
			dest:      dst,
		}
	}

	if in.capturer.captureCancel != nil {
		in.capturer.captureCancel()
	}

	return nil
}

func (in *Input) Stop() error {
	in.started = false
	in.cfg = types.InputConfig{
		Framerate: 1,
		Outputs:   nil,
	}
	in.capturer.captureCancel()
	return nil
}

func (in *Input) processFrame(pix []byte) {
	now := time.Now()

	outs := map[uuid.UUID][]types.UpdateEventOutput{}

	src := &image.RGBA{
		Pix:    pix,
		Stride: in.display.Width() * 4,
		Rect:   image.Rect(0, 0, in.display.Width(), in.display.Height()),
	}

	wg := new(sync.WaitGroup)
	wg.Add(len(in.outputs))
	var outMux sync.Mutex

	err := in.resizer.Convert(in.resized, src)
	if err != nil {
		fmt.Println("error resizing:", err)
		return
	}

	for _, out := range in.outputs {
		out := out
		go func() {
			defer wg.Done()
			rect := image.Rect(out.subLeft/8, out.subTop/8, (out.subLeft+out.subWidth)/8, (out.subTop+out.subHeight)/8)

			sub := in.resized.SubImage(rect)

			out.scaler.Scale(out.dest, out.dest.Bounds(), sub, sub.Bounds(), draw.Over, nil)

			var colors []color.Color

			for i := 0; i < len(out.dest.Pix); i += 4 {
				clr := color.NRGBA{
					R: out.dest.Pix[i],
					G: out.dest.Pix[i+1],
					B: out.dest.Pix[i+2],
					A: out.dest.Pix[i+3],
				}

				colors = append(colors, clr)
			}

			if out.reverse {
				reverse(colors)
			}

			outMux.Lock()
			outs[out.sinkId] = append(outs[out.sinkId], types.UpdateEventOutput{
				OutputId: out.outputId,
				Pix:      colors,
			})
			outMux.Unlock()
		}()
	}

	wg.Wait()

	for sinkId, outs := range outs {

		in.events <- types.UpdateEvent{
			SinkId:  sinkId,
			Outputs: outs,
			Latency: time.Since(now),
		}
	}
}

func reverse[S ~[]E, E any](s S) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}
