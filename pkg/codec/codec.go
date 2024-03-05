package codec

type Codec[E any] interface {
	MarshalBinary(v E) (data []byte, err error)
	UnmarshalBinary(data []byte) (E, error)
}
