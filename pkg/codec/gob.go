package codec

import (
	"bytes"
	"encoding/gob"
)

type GobCodec struct{}

func NewGobCodec(types ...any) Codec {
	for _, typ := range types {
		gob.Register(typ)
	}

	return &GobCodec{}
}

func (m *GobCodec) UnmarshalEvent(b []byte, dest *any) error {
	r := bytes.NewReader(b)
	dec := gob.NewDecoder(r)

	err := dec.Decode(dest)
	if err != nil {
		return err
	}

	return nil
}

func (m *GobCodec) MarshalEvent(e any) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(&e)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
