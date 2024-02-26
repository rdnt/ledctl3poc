package codec

type Codec interface {
	MarshalEvent(e any) ([]byte, error)
	UnmarshalEvent(b []byte, e *any) error
}
