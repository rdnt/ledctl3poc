package audio

import (
	_ "embed"
	"encoding/json"
)

//go:generate go run github.com/atombender/go-jsonschema/cmd/gojsonschema -p audio --tags json -o schema.gen.go schema.json

//go:embed schema.json
var b []byte
var schema map[string]any

func init() {
	_ = json.Unmarshal(b, &schema)
}

func (in *Input) Schema() map[string]any {
	return schema
}

//func (in *Input) ApplyConfig(cfg map[string]any) error {
//	//var config SchemaJson
//	//err := json.Unmarshal(b, &config)
//	//if err != nil {
//	//	return err
//	//}
//
//	//fmt.Printf("applying config: %#v\n", cfg)
//
//	//colors := []color.Color{}
//	//
//	//for _, hex := range cfg["colors"].([]string) {
//	//	clr, err := colorful.Hex(hex)
//	//	if err != nil {
//	//		return err
//	//	}
//	//
//	//	colors = append(colors, clr)
//	//}
//
//	//fmt.Print("===============================")
//	//fmt.Print(colors)
//	//fmt.Print(colors)
//	//fmt.Print(colors)
//
//	//err = WithBlackPoint(cfg["blackPoint"].(float64))(in)
//	//if err != nil {
//	//	fmt.Println("blackpoint error")
//	//	return err
//	//}
//
//	//in.gradient, err = gradient.New(in.colors...)
//	//if err != nil {
//	//	fmt.Println("gradient error")
//	//	return err
//	//}
//
//	//in.events = make(chan source.UpdateEvent, len(in.segments))
//
//	//v.average = make(map[int]sliceewma.MovingAverage, len(v.segments))
//
//	//in.freqMax = ewma.NewMovingAverage(float64(in.windowSize) * 8)
//	//
//	//in.average = make(map[uuid.UUID]pixavg.Average, len(in.segments))
//	//
//	//for _, seg := range in.segments {
//	//	prev := make([]color.Color, seg.Leds)
//	//	for i := 0; i < len(prev); i++ {
//	//		prev[i] = color.RGBA{}
//	//	}
//	//	in.average[seg.Id] = pixavg.New(in.windowSize, prev, 2)
//	//}
//
//	err := in.Stop()
//	if err != nil {
//		fmt.Println("stop error")
//		return err
//	}
//
//	err = in.Start(in.sinkCfg)
//	if err != nil {
//		fmt.Println("start error")
//		return err
//	}
//
//	return nil
//}
