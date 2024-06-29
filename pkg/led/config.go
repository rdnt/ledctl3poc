package led

import (
	_ "embed"
)

//go:generate go-jsonschema -p led --tags json -o sink_schema.gen.go sink_schema.json

//go:embed sink_schema.json
var sinkSchemaBin []byte

func (d *dev) Schema() []byte {
	return sinkSchemaBin
}

//go:generate go-jsonschema -p led --tags json -o output_schema.gen.go output_schema.json

//go:embed output_schema.json
var outputSchemaBin []byte

func (o *output) Schema() []byte {
	return outputSchemaBin
}
