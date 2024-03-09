package event

import (
	"encoding/json"
	"fmt"

	"ledctl3/pkg/codec"
)

type Event interface {
	Type() string
}

type evt struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type JSONCodec struct{}

func (c JSONCodec) MarshalBinary(v Event) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return json.Marshal(evt{
		Type: v.Type(),
		Data: b,
	})
}

func (c JSONCodec) UnmarshalBinary(b []byte) (Event, error) {
	var t evt
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}

	switch t.Type {
	case TypeNodeConnected:
		var e NodeConnected
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeInputConnected:
		var e InputConnected
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeOutputConnected:
		var e OutputConnected
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeInputDisconnected:
		var e InputDisconnected
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeOutputDisconnected:
		var e OutputDisconnected
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeSetSourceConfig:
		var e SetSourceConfig
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeSetSinkConfig:
		var e SetSinkConfig
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeSetSourceConfigCommand:
		var e SetSourceConfigCommand
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeSetSinkConfigCommand:
		var e SetSinkConfigCommand
		err := json.Unmarshal(t.Data, &e)
		return e, err

	default:
		return nil, fmt.Errorf("unknown event type: '%s'", t.Type)
	}
}

func NewJSONCodec() codec.Codec[Event] {
	return JSONCodec{}
}
