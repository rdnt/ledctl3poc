package event

import (
	"image/color"
)

//var Codec codec.Codec[Event]

var Types = []any{
	[]any{},
	map[string]any{},
	color.NRGBA{},
	[]byte{},
	([]byte)(nil),

	NodeConnected{},
	Data{},
	SetSourceActive{},
	SetInputActive{},
	InputConnected{},
	InputDisconnected{},
	OutputConnected{},
	OutputDisconnected{},
}

func init() {
	//Codec = codec.NewGobCodec[Event](
	//	[]any{},
	//	map[string]any{},
	//	color.NRGBA{},
	//	[]byte{},
	//	([]byte)(nil),
	//
	//	NodeConnected{},
	//	Data{},
	//	SetSourceActive{},
	//	SetInputActive{},
	//	InputConnected{},
	//	InputDisconnected{},
	//	OutputConnected{},
	//	OutputDisconnected{},
	//)
}
