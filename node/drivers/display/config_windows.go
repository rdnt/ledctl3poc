package screen

import (
	_ "embed"
)

//go:generate go-jsonschema -p screen --tags json -o source_schema.gen.go source_schema.json

//go:embed source_schema.json
var sourceSchemaBin []byte

func (c *Capturer) Schema() []byte {
	return sourceSchemaBin
}

//go:generate go-jsonschema -p screen --tags json -o input_schema.gen.go input_schema.json

//go:embed input_schema.json
var inputSchemaBin []byte

func (in *input) Schema() []byte {
	return inputSchemaBin
}

//go:generate go-jsonschema -p screen --tags json -o link_schema.gen.go link_schema.json

//go:embed link_schema.json
var linkSchemaBin []byte

func (in *input) LinkSchema() []byte {
	return linkSchemaBin
}
