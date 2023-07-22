package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"ledctl3/pkg/uuid"
	"github.com/samber/lo"

	"ledctl3/event"
	"ledctl3/pkg/broker"
	"ledctl3/registry"
	resolver "ledctl3/registry/pkg"
	sinkdev "ledctl3/sink"
	outputdev "ledctl3/sink/debug"
	"ledctl3/sink/pkg/sinkmdns"
	sourcedev "ledctl3/source"
	audiosrc "ledctl3/source/audio"
	inputdev "ledctl3/source/debug"
	"ledctl3/source/pkg/sourcemdns"
	screensrc "ledctl3/source/screen"
)

func main() {
	time.Sleep(1 * time.Second)

	socket := broker.New[uuid.UUID, event.EventIface]()

	reg, err := registry.New()
	handle(err)

	res, err := resolver.New(reg)
	handle(err)

	err = res.Browse(context.Background())
	handle(err)

	socket.Subscribe(reg.Id(), reg.ProcessEvent)
	go func() {
		for e := range reg.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()

	inputdev1a, err := audiosrc.New()
	handle(err)

	inputdev1b, err := screensrc.New()
	handle(err)

	src1dev, err := sourcedev.New(reg.Id())
	handle(err)

	srv1, err := sourcemdns.New(src1dev)
	handle(err)
	err = srv1.Start()
	handle(err)

	src1dev.AddInput(inputdev1a)
	src1dev.AddInput(inputdev1b)
	socket.Subscribe(src1dev.Id(), src1dev.ProcessEvent)
	go func() {
		for e := range src1dev.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()

	inputdev2a := inputdev.New()
	inputdev2b := inputdev.New()

	src2dev, err := sourcedev.New(reg.Id())
	handle(err)

	srv2, err := sourcemdns.New(src2dev)
	handle(err)
	err = srv2.Start()
	handle(err)

	src2dev.AddInput(inputdev2a)
	src2dev.AddInput(inputdev2b)
	socket.Subscribe(src2dev.Id(), src2dev.ProcessEvent)
	go func() {
		for e := range src2dev.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()

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

	srv3, err := sinkmdns.New(sink1dev)
	handle(err)
	err = srv3.Start()
	handle(err)

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

	srv4, err := sinkmdns.New(sink2dev)
	handle(err)
	err = srv4.Start()
	handle(err)

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

	err = reg.AssistedSetup(inputdev1b.Id())
	handle(err)

	time.Sleep(1 * time.Second)

	cfgs := lo.Values(reg.InputConfigs(inputdev1b.Id()))

	if len(cfgs) == 0 {
		panic("no config")
	}

	cfg := cfgs[0].Cfg
	dis := cfg["displays"].([]map[string]any)
	dis[0]["reverse"] = true
	cfg["displays"] = dis

	reg.UpdateInputConfig(inputdev1b.Id(), cfgs[0].Id, "custom", cfg)

	prof1 := reg.AddProfile("profile1", []registry.ProfileSource{
		//{inputdev1a.OutputId(): {outputdev1a.OutputId(), outputdev2b.OutputId()}},
		//{inputdev2b.OutputId(): {outputdev1b.OutputId()}},
		{
			SourceId: src1dev.Id(),
			Inputs: []registry.ProfileInput{
				{
					InputId: inputdev1b.Id(),
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

	time.Sleep(1 * time.Second)

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
