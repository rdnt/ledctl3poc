package main

import (
	"encoding/json"
	"os"
	"slices"
	"strings"

	"github.com/samber/lo"

	"ledctl3/orchestrator"
)

func getState() (orchestrator.State, error) {
	b, err := os.ReadFile("./tmp/registry.json")
	if err != nil {
		return orchestrator.State{}, err
	}

	var state orchestrator.State
	err = json.Unmarshal(b, &state)
	if err != nil {
		return orchestrator.State{}, err
	}

	return state, nil
}

type Node struct {
	Id           string
	Name         string
	InputsCount  int
	OutputsCount int
	Connected    bool
}

func getNodes(state orchestrator.State) []Node {
	var nodes []Node

	ids := lo.Keys(state.Nodes)
	slices.SortStableFunc(ids, func(a, b uuid.UUID) int {
		cmp := strings.Compare(state.Nodes[a].Name, state.Nodes[b].Name)
		if cmp != 0 {
			return cmp
		}

		return strings.Compare(a.String(), b.String())
	})

	for _, id := range ids {
		node := state.Nodes[id]

		nodes = append(nodes, Node{
			Id:           node.Id.String(),
			Name:         node.Name,
			InputsCount:  len(node.Inputs),
			OutputsCount: len(node.Outputs),
			Connected:    node.Connected,
		})
	}

	return nodes
}
