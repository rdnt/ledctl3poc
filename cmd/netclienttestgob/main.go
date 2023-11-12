package main

import (
	"fmt"
	"ledctl3/event"
	"ledctl3/pkg/codec"
)

func main() {
	cod := codec.NewGobCodec[event.EventIface](
		[]any{},
		map[string]any{},
		event.AssistedSetupEvent{},
		event.AssistedSetupConfigEvent{},
		event.CapabilitiesEvent{},
		event.ConnectEvent{},
		event.DataEvent{},
		event.ListCapabilitiesEvent{},
		event.SetInputConfigEvent{},
		event.SetSinkActiveEvent{},
		event.SetSourceActiveEvent{},
		event.SetSourceIdleEvent{},
	)

	b, err := cod.MarshalEvent(event.SetSourceIdleEvent{Event: event.Event{Type: event.SetSourceIdle}})
	if err != nil {
		panic(err)
	}

	var e event.EventIface
	err = cod.UnmarshalEvent(b, &e)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", e)
}
