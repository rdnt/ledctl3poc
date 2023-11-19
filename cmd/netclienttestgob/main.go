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
		event.AssistedSetup{},
		event.AssistedSetupConfig{},
		event.Capabilities{},
		event.Connect{},
		event.Data{},
		event.ListCapabilities{},
		event.SetInputConfig{},
		event.SetSinkActive{},
		event.SetSourceActive{},
		event.SetSourceIdle{},
	)

	b, err := cod.MarshalEvent(event.SetSourceIdle{Event: event.Event{Type: event.SetSourceIdle}})
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
