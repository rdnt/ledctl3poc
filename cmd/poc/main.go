package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/samber/lo"

	"ledctl3/_registry-old"
	resolver "ledctl3/_registry-old/pkg"
	sinkdev "ledctl3/_sink-old"
	outputdev "ledctl3/_sink-old/debug"
	"ledctl3/_sink-old/pkg/sinkmdns"
	sourcedev "ledctl3/_source-old"
	audiosrc "ledctl3/_source-old/audio"
	inputdev "ledctl3/_source-old/debug"
	"ledctl3/_source-old/pkg/sourcemdns"
	screensrc "ledctl3/_source-old/screen"
	"ledctl3/event"
	"ledctl3/pkg/fsbroker"
)

func main() {
	time.Sleep(1 * time.Second)

	//socket := broker.New[uuid.UUID, event.EventIface]()

	socket := fsbroker.New[event.EventIface]()
	socket.Start()

	reg, err := _registry_old.New()
	handle(err)

	res, err := resolver.New(reg)
	handle(err)

	err = res.Browse(context.Background())
	handle(err)

	socket.Subscribe(reg.Id(), reg.ProcessEvent)
	go func() {
		for e := range reg.Messages() {
			socket.Publish(e.Address(), e)
		}
	}()

	inputdev1a, err := audiosrc.New()
	handle(err)

	inputdev1b := inputdev.New()

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
		for e := range src1dev.Messages() {
			socket.Publish(e.Address(), e)
		}
	}()

	//ins, err := screenProv.Inputs()
	//handle(err)

	//inputdev2a := inputdev.New()
	//inputdev2b := inputdev.New()

	src2dev, err := sourcedev.New(reg.Id())
	handle(err)

	srv2, err := sourcemdns.New(src2dev)
	handle(err)
	err = srv2.Start()
	handle(err)

	//for _, d := range ins {
	//	src2dev.AddInput(d)
	//}
	screenProv, err := screensrc.New(src2dev)
	handle(err)
	screenProv.Start()
	//src2dev.AddInput(inputdev2a)
	//src2dev.AddInput(inputdev2b)
	socket.Subscribe(src2dev.Id(), src2dev.ProcessEvent)
	go func() {
		for e := range src2dev.Messages() {
			socket.Publish(e.Address(), e)
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
		for e := range sink1dev.Messages() {
			socket.Publish(e.Address(), e)
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
		for e := range sink2dev.Messages() {
			socket.Publish(e.Address(), e)
		}
	}()

	srv4, err := sinkmdns.New(sink2dev)
	handle(err)
	err = srv4.Start()
	handle(err)

	//////////////////////////

	//input1a := source.NewInput(inputdev1a.Id(), "input1a")
	//input1b := source.NewInput(inputdev1b.Id(), "input1b")
	//
	//source1 := source.NewSource(src1dev.Id(), "source1", map[uuid.UUID]*source.Capturer{
	//	input1a.Id(): input1a, input1b.Id(): input1b,
	//})
	//
	//_ = source1
	////err := reg.AddSource(source1)
	////handle(err)
	//
	//input2a := source.NewInput(inputdev2a.Id(), "input2a")
	//input2b := source.NewInput(inputdev2b.Id(), "input2b")
	//
	//source2 := source.NewSource(src2dev.Id(), "source2", map[uuid.UUID]*source.Capturer{
	//	input2a.Id(): input2a, input2b.Id(): input2b,
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
	//sink1 := sink.NewSink(sink1dev.Id(), "sink1", map[uuid.UUID]*sink.Output{
	//	output1a.Id(): output1a, output1b.Id(): output1b,
	//})
	//
	//_ = sink1
	////err = reg.AddSink(sink1)
	////handle(err)
	//
	//output2a := sink.NewOutput(uuid.New(), "output2a", 16)
	//output2b := sink.NewOutput(uuid.New(), "output2b", 32)
	//
	//sink2 := sink.NewSink(sink2dev.Id(), "sink2", map[uuid.UUID]*sink.Output{
	//	output2a.Id(): output2a, output2b.Id(): output2b,
	//})
	//
	//_ = sink2

	//sources := reg.Sources()
	//sinks := reg.Sinks()

	//err = reg.AddSink(sink2)
	//handle(err)

	//////////////////////////

	time.Sleep(3 * time.Second)

	ins, err := screenProv.Inputs()
	handle(err)

	fmt.Println("Assisted setup:", ins[0])

	err = reg.AssistedSetup(ins[0].Id())
	handle(err)

	time.Sleep(1 * time.Second)

	cfgs := lo.Values(reg.InputConfigs(ins[0].Id()))

	if len(cfgs) == 0 {
		panic("no config")
	}

	cfgs[0].Cfg["reverse"] = true

	reg.UpdateInputConfig(ins[0].Id(), cfgs[0].Id, "custom", cfgs[0].Cfg)

	prof1, _ := reg.AddProfile("profile1", []_registry_old.ProfileSource{
		//{inputdev1a.Id(): {outputdev1a.Id(), outputdev2b.Id()}},
		//{inputdev2b.Id(): {outputdev1b.Id()}},
		{
			SourceId: src2dev.Id(),
			Inputs: []_registry_old.ProfileInput{
				{
					InputId: ins[0].Id(),
					Sinks: []_registry_old.ProfileSink{
						{
							SinkId: sink2dev.Id(),
							Outputs: []_registry_old.ProfileOutput{
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

		//{inputdev1a.Id(): {outputdev2b.Id()}}, // audio
		//{inputdev1b.Id(): {outputdev2b.Id()}}, // video
	})

	//prof2 := reg.CreateProfile("profile2", []map[uuid.UUID][]uuid.UUID{
	//	{input1a.Id(): {output1a.Id(), output2b.Id()}},
	//	{input2b.Id(): {output1b.Id()}},
	//})

	//fmt.Println("==== registry ===")
	//fmt.Println(reg)
	//fmt.Print("========================== \n\n\n")

	err = reg.SelectProfile(prof1.Id, true)
	handle(err)

	time.Sleep(5 * time.Second)

	err = reg.SelectProfile(prof1.Id, false)
	handle(err)

	time.Sleep(5 * time.Second)

	err = reg.SelectProfile(prof1.Id, true)
	handle(err)

	time.Sleep(5 * time.Second)

	err = reg.SelectProfile(prof1.Id, false)
	handle(err)

	time.Sleep(5 * time.Second)

	//err = reg.ConfigureInput(inputdev1a.Id(), map[string]any{
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
	//err = reg.EnableProfile(prof2.Id)
	//handle(err)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c
}

func handle(err error) {
	return
	if err != nil {
		panic(err)
	}
}
