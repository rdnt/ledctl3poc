package req

import (
	"encoding/json"
	"fmt"

	"ledctl3/pkg/codec"
)

type Request interface {
	Type() string
}

type Response interface {
	Type() string
}

type req struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type JSONCodec struct{}

func (c JSONCodec) MarshalBinary(v Request) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	return json.Marshal(req{
		Type: v.Type(),
		Data: b,
	})
}

func (c JSONCodec) UnmarshalBinary(b []byte) (Request, error) {
	var t req
	if err := json.Unmarshal(b, &t); err != nil {
		return nil, err
	}

	switch t.Type {
	case TypeSetSourceConfig:
		var e SetSourceConfig
		err := json.Unmarshal(t.Data, &e)
		return e, err

	case TypeSetSinkConfig:
		var e SetSinkConfig
		err := json.Unmarshal(t.Data, &e)
		return e, err

	default:
		return nil, fmt.Errorf("unknown request type: '%s'", t.Type)
	}
}

func NewJSONCodec() codec.Codec[Request] {
	return JSONCodec{}
}
