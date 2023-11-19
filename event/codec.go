package event

import "ledctl3/pkg/codec"

var Codec codec.Codec[Event]

func init() {
	Codec = codec.NewGobCodec[Event](
		[]any{},
		map[string]any{},
		AssistedSetup{},
		AssistedSetupConfig{},
		Capabilities{},
		Connect{},
		Data{},
		ListCapabilities{},
		SetInputConfig{},
		SetSinkActive{},
		SetSourceActive{},
		SetSourceIdle{},
		InputConnected{},
		InputDisconnected{},
		OutputConnected{},
		OutputDisconnected{},
	)

}
