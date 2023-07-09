package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/google/uuid"

	"ledctl3/event"
	"ledctl3/pkg/broker"
	"ledctl3/registry"
	"ledctl3/registry/types/sink"
	"ledctl3/registry/types/source"
	sourcedev "ledctl3/source"
)

func main() {
	time.Sleep(1 * time.Second)

	reg := registry.New()

	socket := broker.New[uuid.UUID, event.EventIface]()

	src1dev := sourcedev.New(nil)
	socket.Subscribe(src1dev.Id(), src1dev.ProcessEvent)
	go func() {
		for e := range src1dev.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()

	src2dev := sourcedev.New(nil)
	socket.Subscribe(src2dev.Id(), src2dev.ProcessEvent)
	go func() {
		for e := range src1dev.Events() {
			socket.Publish(e.DeviceId(), e)
		}
	}()

	//////////////////////////

	input1a := source.NewInput(uuid.New(), "input1a")
	input1b := source.NewInput(uuid.New(), "input1b")

	source1 := source.NewSource(src1dev.Id(), "source1", map[uuid.UUID]*source.Input{
		input1a.Id(): input1a, input1b.Id(): input1b,
	}, func(e event.EventIface) error {
		//fmt.Printf("@@@ send %s (%s) %s\n", src1dev.Id(), e.DeviceId(), e.Type())
		socket.Publish(e.DeviceId(), e)
		return nil
	}, func() <-chan event.EventIface {
		evts := make(chan event.EventIface)
		socket.Subscribe(src1dev.Id(), func(e event.EventIface) {
			evts <- e
		})
		return evts
	})

	err := reg.AddSource(source1)
	handle(err)

	input2a := source.NewInput(uuid.New(), "input2a")
	input2b := source.NewInput(uuid.New(), "input2b")

	source2 := source.NewSource(src2dev.Id(), "source2", map[uuid.UUID]*source.Input{
		input2a.Id(): input2a, input2b.Id(): input2b,
	}, func(e event.EventIface) error {
		//fmt.Printf("@@@ send %s (%s) %s\n", src2dev.Id(), e.DeviceId(), e.Type())
		socket.Publish(e.DeviceId(), e)
		return nil
	}, func() <-chan event.EventIface {
		evts := make(chan event.EventIface)
		socket.Subscribe(src2dev.Id(), func(e event.EventIface) {
			evts <- e
		})
		return evts
	})

	err = reg.AddSource(source2)
	handle(err)

	//////////////////////////

	output1a := sink.NewOutput(uuid.New(), "output1a")
	output1b := sink.NewOutput(uuid.New(), "output1b")

	sink1id := uuid.New()
	// TODO: actual device uuid
	sink1 := sink.NewSink(sink1id, "sink1", map[uuid.UUID]*sink.Output{
		output1a.Id(): output1a, output1b.Id(): output1b,
	}, func(e event.EventIface) error {
		//fmt.Printf("~~~ send %s (%s) %s\n", sink1id, e.DeviceId(), e.Type())
		socket.Publish(e.DeviceId(), e)
		return nil
	}, func() <-chan event.EventIface {
		evts := make(chan event.EventIface)
		socket.Subscribe(sink1id, func(e event.EventIface) {
			evts <- e
		})
		return evts
	})

	socket.Subscribe(sink1id, func(e event.EventIface) {
		fmt.Println("### debug sink1 recv", e.Type())
	})

	err = reg.AddSink(sink1)
	handle(err)

	output2a := sink.NewOutput(uuid.New(), "output2a")
	output2b := sink.NewOutput(uuid.New(), "output2b")

	sink2id := uuid.New()
	// TODO: actual device uuid
	sink2 := sink.NewSink(sink2id, "sink2", map[uuid.UUID]*sink.Output{
		output2a.Id(): output2a, output2b.Id(): output2b,
	}, func(e event.EventIface) error {
		//fmt.Printf("~~~ send %s (%s) %s\n", sink2id, e.DeviceId(), e.Type())
		socket.Publish(e.DeviceId(), e)
		return nil
	}, func() <-chan event.EventIface {
		evts := make(chan event.EventIface)
		socket.Subscribe(sink2id, func(e event.EventIface) {
			evts <- e
		})
		return evts
	})

	socket.Subscribe(sink2id, func(e event.EventIface) {
		fmt.Println("### debug sink2 recv", e.Type())
	})

	err = reg.AddSink(sink2)
	handle(err)

	//////////////////////////

	prof1 := reg.AddProfile("profile1", []map[uuid.UUID][]uuid.UUID{
		{input1a.Id(): {output1a.Id(), output2b.Id()}},
		{input2b.Id(): {output1b.Id()}},
	})

	prof2 := reg.AddProfile("profile2", []map[uuid.UUID][]uuid.UUID{
		{input1a.Id(): {output1a.Id(), output2b.Id()}},
		{input2b.Id(): {output1b.Id()}},
	})

	fmt.Println("=== REGISTRY ===")
	fmt.Println(reg)
	fmt.Print("========================== \n\n\n")

	time.Sleep(1 * time.Second)

	err = reg.SelectProfile(prof1.Id)
	handle(err)

	time.Sleep(5 * time.Second)

	err = reg.SelectProfile(prof2.Id)
	handle(err)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}
