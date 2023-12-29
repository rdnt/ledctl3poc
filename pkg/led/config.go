package led

import (
	_ "embed"
)

//go install github.com/atombender/go-jsonschema@latest
//go:generate go-jsonschema -p led --tags json -o schema.gen.go schema.json

//go:embed schema.json
var schema []byte

func (d *Device) Schema() ([]byte, error) {
	return schema, nil
}
