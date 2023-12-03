package codec

import (
	"bytes"
	"encoding/gob"
)

type GobCodec[E any] struct{}

func NewGobCodec[E any](types ...any) Codec[E] {
	for _, typ := range types {
		gob.Register(typ)
	}

	return &GobCodec[E]{}
}

func (m *GobCodec[E]) UnmarshalEvent(b []byte, dest *E) error {
	r := bytes.NewReader(b)
	dec := gob.NewDecoder(r)

	err := dec.Decode(dest)
	if err != nil {
		return err
	}

	return nil
}

func (m *GobCodec[E]) MarshalEvent(e E) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	err := enc.Encode(&e)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
