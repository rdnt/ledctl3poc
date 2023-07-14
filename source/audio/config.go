package audio

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:generate go run github.com/atombender/go-jsonschema/cmd/gojsonschema -p audio --tags json -o schema.gen.go schema.json

//go:embed schema.json
var b []byte
var schema map[string]any

func init() {
	_ = json.Unmarshal(b, &schema)
}

func (v *Visualizer) Schema() map[string]any {
	return schema
}

func (v *Visualizer) ApplyConfig(cfg map[string]any) error {
	//var config SchemaJson
	//err := json.Unmarshal(b, &config)
	//if err != nil {
	//	return err
	//}

	fmt.Printf("applying config: %#v\n", cfg)

	return nil
}
