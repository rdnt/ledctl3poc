package main

import (
	"encoding/json"
	"os"

	"ledctl3/internal/registry"
)

func getState() (registry.PersistentState, error) {
	b, err := os.ReadFile("./tmp/registry.json")
	if err != nil {
		return registry.PersistentState{}, err
	}

	var state registry.PersistentState
	err = json.Unmarshal(b, &state)
	if err != nil {
		return registry.PersistentState{}, err
	}

	return state, nil
}
