package event

import (
	"image/color"

	"ledctl3/pkg/codec"
)

var Codec codec.Codec[Event]

func init() {
	Codec = codec.NewGobCodec[Event](
		[]any{},
		map[string]any{},
		NodeConnected{},
		Data{},
		SetSourceActive{},
		SetInputActive{},
		InputConnected{},
		InputDisconnected{},
		OutputConnected{},
		OutputDisconnected{},
		color.NRGBA{},
		SetSourceConfig{},
		SetSinkConfig{},
		[]byte{},
		([]byte)(nil),
	)
}
