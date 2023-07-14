package registry

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:generate go run github.com/atombender/go-jsonschema/cmd/gojsonschema -p registry --tags json -o inputschema.gen.go input-schema.json

func ParseConfig(b []byte) error {
	var config map[string]json.RawMessage
	err := json.Unmarshal(b, &config)
	if err != nil {
		return err
	}

	for _, b := range config {
		var optJson ConfigOption
		err = json.Unmarshal(b, &optJson)
		if err != nil {
			return err
		}

		err = parseOption(optJson.Type, b)
		if err != nil {
			return err
		}

		//fmt.Printf("%#v\n", optJson)
	}

	return nil
}

func parseOption(typ string, b []byte) error {
	switch typ {
	case "boolean":
		var opt BooleanOption
		err := json.Unmarshal(b, &opt)
		if err != nil {
			return err
		}

		//fmt.Printf("%#v\n", opt)
	case "string":
		var opt StringOption
		err := json.Unmarshal(b, &opt)
		if err != nil {
			return err
		}

		fmt.Printf("%#v\n", opt)
	case "integer":
		var opt IntegerOption
		err := json.Unmarshal(b, &opt)
		if err != nil {
			return err
		}

		//fmt.Printf("%#v\n", opt)
	case "float":
		var opt FloatOption
		err := json.Unmarshal(b, &opt)
		if err != nil {
			return err
		}

		//fmt.Printf("%#v\n", opt)
	case "array":
		var opt ArrayOption
		err := json.Unmarshal(b, &opt)
		if err != nil {
			return err
		}

		b, err := json.Marshal(opt.Items)
		if err != nil {
			return err
		}

		var optJson ConfigOption
		err = json.Unmarshal(b, &optJson)
		if err != nil {
			return err
		}

		var params json.RawMessage
		err = json.Unmarshal(b, &params)
		if err != nil {
			return err
		}

		//fmt.Printf("%#v\n", opt)

		err = parseOption(optJson.Type, params)
		if err != nil {
			return err
		}
	case "object":
		var opt ObjectOption
		err := json.Unmarshal(b, &opt)
		if err != nil {
			return err
		}

		b, err := json.Marshal(opt.Properties)
		if err != nil {
			return err
		}

		var optJson map[string]json.RawMessage
		err = json.Unmarshal(b, &optJson)
		if err != nil {
			return err
		}

		var config map[string]json.RawMessage
		err = json.Unmarshal(b, &config)
		if err != nil {
			return err
		}

		for _, b := range config {
			var optJson ConfigOption
			err = json.Unmarshal(b, &optJson)
			if err != nil {
				return err
			}

			err = parseOption(optJson.Type, b)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

type Config []ConfigOption

type ConfigOption struct {
	Type string `json:"type"`
	//Name string `json:"name"`
}
