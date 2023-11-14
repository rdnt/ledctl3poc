package event

import "ledctl3/pkg/codec"

var Codec codec.Codec[EventIface]

func init() {
	Codec = codec.NewGobCodec[EventIface](
		[]any{},
		map[string]any{},
		AssistedSetupEvent{},
		AssistedSetupConfigEvent{},
		CapabilitiesEvent{},
		ConnectEvent{},
		DataEvent{},
		ListCapabilitiesEvent{},
		SetInputConfigEvent{},
		SetSinkActiveEvent{},
		SetSourceActiveEvent{},
		SetSourceIdleEvent{},
	)
}
