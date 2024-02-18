package main

import (
	"encoding/json"
	"os"

	"ledctl3/registry"
)

func getState() (registry.State, error) {
	b, err := os.ReadFile("./tmp/registry.json")
	if err != nil {
		return registry.State{}, err
	}

	var state registry.State
	err = json.Unmarshal(b, &state)
	if err != nil {
		return registry.State{}, err
	}

	return state, nil
}
