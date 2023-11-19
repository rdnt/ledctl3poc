package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"

	"ledctl3/event"
	"ledctl3/pkg/fsbroker"
	"ledctl3/pkg/uuid"
)

func main2() {
	socket := fsbroker.New[event.EventIface]()
	socket.Start()

	socket.Subscribe(uuid.Nil, func(e event.EventIface) {
		fmt.Println("EVENT", e)
	})

	time.Sleep(500 * time.Millisecond)

	socket.Publish(uuid.Nil, event.AssistedSetup{
		Event: event.Event{
			Type: event.AssistedSetup,
			Addr: uuid.Nil,
		},
		InputId: uuid.New(),
	})

	select {}
}

func main() {
	gob.Register(event.AssistedSetup{})
	gob.Register(event.AssistedSetupConfig{})
	gob.Register(event.Capabilities{})
	gob.Register(event.Connect{})
	gob.Register(event.Data{})
	gob.Register(event.ListCapabilities{})
	gob.Register(event.SetInputConfig{})
	gob.Register(event.SetSinkActive{})
	gob.Register(event.SetSourceActive{})
	gob.Register(event.SetSourceIdle{})

	e := event.AssistedSetup{
		Event: event.Event{
			Type: event.AssistedSetup,
			Addr: uuid.New(),
		},
		InputId: uuid.New(),
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	eee := event.EventIface(e)

	err := enc.Encode(&eee) // reference to iface!
	if err != nil {
		panic(err)
	}

	dec := gob.NewDecoder(&b)
	var e2 event.EventIface
	//e2 := reflect.New(reflect.TypeOf(event.EventIface(event.Data{})))
	err = dec.Decode(&e2)
	if err != nil {
		panic(err)
	}

	fmt.Println(e2)
}
