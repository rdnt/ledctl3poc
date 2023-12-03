package codec

type Codec[E any] interface {
	MarshalEvent(e E) ([]byte, error)
	UnmarshalEvent(b []byte, e *E) error
}
