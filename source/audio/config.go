package audio

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image/color"

	"github.com/VividCortex/ewma"
	"github.com/google/uuid"
	"github.com/lucasb-eyer/go-colorful"

	"ledctl3/pkg/gradient"
	"ledctl3/pkg/pixavg"
)

//go:generate go run github.com/atombender/go-jsonschema/cmd/gojsonschema -p audio --tags json -o schema.gen.go schema.json

//go:embed schema.json
var b []byte
var schema map[string]any

func init() {
	_ = json.Unmarshal(b, &schema)
}

func (a *AudioCapture) Schema() map[string]any {
	return schema
}

func (a *AudioCapture) ApplyConfig(cfg map[string]any) error {
	//var config SchemaJson
	//err := json.Unmarshal(b, &config)
	//if err != nil {
	//	return err
	//}

	fmt.Printf("applying config: %#v\n", cfg)

	colors := []color.Color{}

	for _, hex := range cfg["colors"].([]string) {
		clr, err := colorful.Hex(hex)
		if err != nil {
			return err
		}

		colors = append(colors, clr)
	}

	fmt.Print("===============================")
	fmt.Print(colors)
	fmt.Print(colors)
	fmt.Print(colors)

	err := WithColors(
		colors...,
	)(a)
	if err != nil {
		return err
	}
	err = WithWindowSize(cfg["windowSize"].(int))(a)
	if err != nil {
		fmt.Println("windowsize error")
		return err
	}
	//err = WithBlackPoint(cfg["blackPoint"].(float64))(a)
	//if err != nil {
	//	fmt.Println("blackpoint error")
	//	return err
	//}

	a.gradient, err = gradient.New(a.colors...)
	if err != nil {
		fmt.Println("gradient error")
		return err
	}

	//a.events = make(chan source.UpdateEvent, len(a.segments))

	//v.average = make(map[int]sliceewma.MovingAverage, len(v.segments))

	a.freqMax = ewma.NewMovingAverage(float64(a.windowSize) * 8)

	a.average = make(map[uuid.UUID]pixavg.Average, len(a.segments))

	for _, seg := range a.segments {
		prev := make([]color.Color, seg.Leds)
		for i := 0; i < len(prev); i++ {
			prev[i] = color.RGBA{}
		}
		a.average[seg.OutputId] = pixavg.New(a.windowSize, prev, 2)
	}

	err = a.Stop()
	if err != nil {
		fmt.Println("stop error")
		return err
	}

	err = a.Start(a.sinkCfg)
	if err != nil {
		fmt.Println("start error")
		return err
	}

	return nil
}
