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

	socket.Publish(uuid.Nil, event.AssistedSetupEvent{
		Event: event.Event{
			Type: event.AssistedSetup,
			Addr: uuid.Nil,
		},
		InputId: uuid.New(),
	})

	select {}
}

func main() {
	gob.Register(event.AssistedSetupEvent{})
	gob.Register(event.AssistedSetupConfigEvent{})
	gob.Register(event.CapabilitiesEvent{})
	gob.Register(event.ConnectEvent{})
	gob.Register(event.DataEvent{})
	gob.Register(event.ListCapabilitiesEvent{})
	gob.Register(event.SetInputConfigEvent{})
	gob.Register(event.SetSinkActiveEvent{})
	gob.Register(event.SetSourceActiveEvent{})
	gob.Register(event.SetSourceIdleEvent{})

	e := event.AssistedSetupEvent{
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
	//e2 := reflect.New(reflect.TypeOf(event.EventIface(event.DataEvent{})))
	err = dec.Decode(&e2)
	if err != nil {
		panic(err)
	}

	fmt.Println(e2)
}
