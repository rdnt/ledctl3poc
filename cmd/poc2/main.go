package main

import (
	"fmt"
	"image/color"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"ledctl3/event"
	"ledctl3/pkg/broker"
	"ledctl3/registry"
	sinkdev "ledctl3/sink"
	outputdev "ledctl3/sink/debug"
	sourcedev "ledctl3/source"
	audiosrc "ledctl3/source/audio"
	inputdev "ledctl3/source/debug"
	videosrc "ledctl3/source/video"
)

func main() {
	time.Sleep(1 * time.Second)

	socket := broker.New[uuid.UUID, event.EventIface]()

	reg := registry.New()
	socket.Subscribe(reg.Id(), reg.ProcessEvent)
	go func() {
		for e := range reg.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()

	inputdev1a, err := audiosrc.New(
		audiosrc.WithColors(
			color.RGBA{0x4a, 0x15, 0x24, 255},
			color.RGBA{0x06, 0x53, 0x94, 255},
			color.RGBA{0x00, 0xb5, 0x85, 255},
			color.RGBA{0xd6, 0x00, 0xa4, 255},
			color.RGBA{0xff, 0x00, 0x4c, 255},
		),
		audiosrc.WithWindowSize(10),
		audiosrc.WithBlackPoint(0),
	)
	handle(err)

	inputdev1b := videosrc.New()

	src1dev := sourcedev.New(reg.Id())
	src1dev.AddInput(inputdev1a)
	src1dev.AddInput(inputdev1b)
	socket.Subscribe(src1dev.Id(), src1dev.ProcessEvent)
	go func() {
		for e := range src1dev.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()
	src1dev.Connect()

	inputdev2a := inputdev.New()
	inputdev2b := inputdev.New()

	src2dev := sourcedev.New(reg.Id())
	src2dev.AddInput(inputdev2a)
	src2dev.AddInput(inputdev2b)
	socket.Subscribe(src2dev.Id(), src2dev.ProcessEvent)
	go func() {
		for e := range src2dev.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()
	src2dev.Connect()

	//////////////////////////

	outputdev1a := outputdev.New(5)
	outputdev1b := outputdev.New(10)

	sink1dev := sinkdev.New(reg.Id())
	sink1dev.AddOutput(outputdev1a)
	sink1dev.AddOutput(outputdev1b)
	socket.Subscribe(sink1dev.Id(), sink1dev.ProcessEvent)
	go func() {
		for e := range sink1dev.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()
	sink1dev.Connect()

	outputdev2a := outputdev.New(20)
	outputdev2b := outputdev.New(120)

	sink2dev := sinkdev.New(reg.Id())
	sink2dev.AddOutput(outputdev2a)
	sink2dev.AddOutput(outputdev2b)
	socket.Subscribe(sink2dev.Id(), sink2dev.ProcessEvent)
	go func() {
		for e := range sink2dev.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()
	sink2dev.Connect()

	//////////////////////////

	//input1a := source.NewInput(inputdev1a.OutputId(), "input1a")
	//input1b := source.NewInput(inputdev1b.OutputId(), "input1b")
	//
	//source1 := source.NewSource(src1dev.OutputId(), "source1", map[uuid.UUID]*source.Input{
	//	input1a.OutputId(): input1a, input1b.OutputId(): input1b,
	//})
	//
	//_ = source1
	////err := reg.AddSource(source1)
	////handle(err)
	//
	//input2a := source.NewInput(inputdev2a.OutputId(), "input2a")
	//input2b := source.NewInput(inputdev2b.OutputId(), "input2b")
	//
	//source2 := source.NewSource(src2dev.OutputId(), "source2", map[uuid.UUID]*source.Input{
	//	input2a.OutputId(): input2a, input2b.OutputId(): input2b,
	//})
	//
	//_ = source2
	////err = reg.AddSource(source2)
	////handle(err)
	//
	////////////////////////////
	//
	//output1a := sink.NewOutput(uuid.New(), "output1a", 4)
	//output1b := sink.NewOutput(uuid.New(), "output1b", 8)
	//
	//sink1 := sink.NewSink(sink1dev.OutputId(), "sink1", map[uuid.UUID]*sink.Output{
	//	output1a.OutputId(): output1a, output1b.OutputId(): output1b,
	//})
	//
	//_ = sink1
	////err = reg.AddSink(sink1)
	////handle(err)
	//
	//output2a := sink.NewOutput(uuid.New(), "output2a", 16)
	//output2b := sink.NewOutput(uuid.New(), "output2b", 32)
	//
	//sink2 := sink.NewSink(sink2dev.OutputId(), "sink2", map[uuid.UUID]*sink.Output{
	//	output2a.OutputId(): output2a, output2b.OutputId(): output2b,
	//})
	//
	//_ = sink2

	//sources := reg.Sources()
	//sinks := reg.Sinks()

	//err = reg.AddSink(sink2)
	//handle(err)

	//////////////////////////

	time.Sleep(1 * time.Second)

	err = reg.AssistedSetup(inputdev1a.Id())
	handle(err)

	time.Sleep(1 * time.Second)

	cfgs := lo.Values(reg.InputConfigs(inputdev1a.Id()))
	fmt.Println("@@@", cfgs)

	prof1 := reg.AddProfile("profile1", []registry.ProfileSource{
		//{inputdev1a.OutputId(): {outputdev1a.OutputId(), outputdev2b.OutputId()}},
		//{inputdev2b.OutputId(): {outputdev1b.OutputId()}},
		{
			SourceId: src1dev.Id(),
			Inputs: []registry.ProfileInput{
				{
					InputId: inputdev1a.Id(),
					Sinks: []registry.ProfileSink{
						{
							SinkId: sink2dev.Id(),
							Outputs: []registry.ProfileOutput{
								{
									OutputId:      outputdev2b.Id(),
									InputConfigId: cfgs[0].Id,
								},
							},
						},
					},
				},
			},
		},

		//{inputdev1a.OutputId(): {outputdev2b.OutputId()}}, // audio
		//{inputdev1b.OutputId(): {outputdev2b.OutputId()}}, // video
	})

	//prof2 := reg.AddProfile("profile2", []map[uuid.UUID][]uuid.UUID{
	//	{input1a.OutputId(): {output1a.OutputId(), output2b.OutputId()}},
	//	{input2b.OutputId(): {output1b.OutputId()}},
	//})

	//fmt.Println("==== registry ===")
	//fmt.Println(reg)
	//fmt.Print("========================== \n\n\n")

	err = reg.SelectProfile(prof1.Id)
	handle(err)

	time.Sleep(2 * time.Second)

	//err = reg.ConfigureInput(inputdev1a.OutputId(), map[string]any{
	//	"colors": []string{
	//		"#4a1524",
	//		"#065394",
	//		"#00b585",
	//		"#d600a4",
	//		"#ff004c",
	//	},
	//	"windowSize": 40,
	//	"blackPoint": 0.2,
	//})
	//handle(err)

	//time.Sleep(5 * time.Second)
	//
	//err = reg.SelectProfile(prof2.OutputId)
	//handle(err)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}
