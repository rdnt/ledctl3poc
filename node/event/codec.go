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
	case TypeAuthenticate:
		var e Authenticate
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeCapabilities:
		var e Capabilities
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

	case TypeConfigureSource:
		var e ConfigureSource
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeConfigureSink:
		var e ConfigureSink
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeData:
		var e Data
		err := json.Unmarshal(t.Data, &e)
		return e, err

	default:
		return nil, fmt.Errorf("unknown event type: '%s'", t.Type)
	}
}

func NewJSONCodec() codec.Codec[Event] {
	return JSONCodec{}
}
