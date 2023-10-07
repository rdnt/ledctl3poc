package screen

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/image/draw"

	"ledctl3/pkg/uuid"

	types2 "ledctl3/pkg/screencapture/types"
	"ledctl3/source/types"
)

type Input struct {
	id      uuid.UUID
	events  chan types.UpdateEvent
	display types2.Display
	outputs map[uuid.UUID]outputCaptureConfig
}

func (in *Input) Events() <-chan types.UpdateEvent {
	return in.events
}

type outputCaptureConfig struct {
	id      uuid.UUID
	sinkId  uuid.UUID
	leds    int
	reverse bool
	scaler  draw.Scaler
}

func (in *Input) AssistedSetup() (map[string]any, error) {
	cfg := map[string]any{
		"width":     in.display.Width(),
		"height":    in.display.Height(),
		"left":      in.display.X(),
		"top":       in.display.Y(),
		"framerate": 60,
	}

	return cfg, nil
}

func (in *Input) Id() uuid.UUID {
	return in.id
}

func (in *Input) Start(cfg types.SinkConfig) error {
	in.outputs = make(map[uuid.UUID]outputCaptureConfig)

	width := in.display.Width()
	height := in.display.Width()

	for _, sinkCfg := range cfg.Sinks {
		for _, out := range sinkCfg.Outputs {
			reverse, _ := out.Config["reverse"].(bool)

			in.outputs[out.Id] = outputCaptureConfig{
				id:      out.Id,
				sinkId:  sinkCfg.Id,
				leds:    out.Leds,
				reverse: reverse,
				scaler:  draw.BiLinear.NewScaler(width, height, width/80, height/80),
			}
		}
	}

	fmt.Printf("## starting screen capture with outputs config %#v\n", in.outputs)

	in.startCapture()

	//err := in.displays.Start()
	//if err != nil {
	//	return err
	//}

	return nil
}

func (in *Input) startCapture() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		frames := in.display.Capture(ctx, 60) // TODO: framerate

		for frame := range frames {
			fmt.Println(in.display.Resolution())

			go in.processFrame(in.display, frame)
		}
	}()

	//displayConfigs, err := in.matchDisplays(in.displays)
	//if err != nil {
	//	return err
	//}
	//
	//in.scalers = make(map[int]draw.Scaler)

	//for _, cfg := range displayConfigs {
	//	for _, seg := range cfg.Segments {
	//		rect := image.Rect(seg.From.X, seg.From.Y, seg.To.X, seg.To.Y)
	//
	//		// TODO: only allow cube (Dx == Dy) if segment is only 1 led
	//
	//		var width, height int
	//
	//		if rect.Dx() > rect.Dy() {
	//			// horizontal
	//			width = seg.Leds
	//			height = 2
	//		} else {
	//			// vertical
	//			width = 2
	//			height = seg.Leds
	//		}
	//
	//		in.scalers[seg.Id] = draw.BiLinear.NewScaler(width, height, cfg.Width, cfg.Height)
	//	}
	//}

	//var wg sync.WaitGroup
	//wg.Add(len(displays))
	//
	//for _, d := range displays {
	//	//cfg := displayConfigs[d.Id()]
	//
	//	go func(d types2.Display) {
	//		defer wg.Done()
	//		frames := d.Capture(ctx, 60) // TODO: framerate
	//
	//		for frame := range frames {
	//			fmt.Println(d.Resolution())
	//
	//			go in.processFrame(d, frame)
	//		}
	//
	//		cancel()
	//	}(d)
	//}
	//
	//wg.Wait()

	return nil
}

func (in *Input) processFrame(d types2.Display, pix []byte) {
	now := time.Now()

	//src := &image.RGBA{
	//	Pix:    pix,
	//	Stride: d.Width() * 4,
	//	Rect:   image.Rect(0, 0, d.Width(), d.Height()),
	//}

	//segs := make([]visualizer.Segment, len(cfg.Segments))

	for _, _ = range in.outputs {
		//rect := image.Rect(seg.From.X, seg.From.Y, seg.To.X, seg.To.Y)
		//
		//sub := src.SubImage(rect)
		//
		//var dst *image.RGBA
		//
		//if rect.Dx() > rect.Dy() {
		//	// horizontal
		//	dst = image.NewRGBA(image.Rect(0, 0, seg.Leds, 1))
		//} else {
		//	// vertical
		//	dst = image.NewRGBA(image.Rect(0, 0, 1, seg.Leds))
		//}

		//v.scalers[seg.Id].Scale(dst, dst.Bounds(), sub, sub.Bounds(), draw.Over, nil)

		//colors := []color.Color{}

		//for i := 0; i < len(dst.Pix); i += 4 {
		//	clr, _ := colorful.MakeColor(color.NRGBA{
		//		R: dst.Pix[i],
		//		G: dst.Pix[i+1],
		//		B: dst.Pix[i+2],
		//		A: dst.Pix[i+3],
		//	})
		//
		//	colors = append(colors, clr)
		//}

		//if seg.Reverse {
		//	reverse(colors)
		//}

		//segs[i] = visualizer.Segment{
		//	Id:  seg.Id,
		//	Pix: colors,
		//}

	}

	//wg.Wait()

	in.events <- types.UpdateEvent{
		//Segments: segs,
		Latency: time.Since(now),
	}
}

func (in *Input) Stop() error {
	return nil
}
