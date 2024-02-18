package ports

import (
	"ledctl3/node/common"
	"ledctl3/pkg/uuid"
)

type Ports struct {
	inputs  map[uuid.UUID]common.Input
	outputs map[uuid.UUID]common.Output
}

func New() *Ports {
	return &Ports{
		inputs:  make(map[uuid.UUID]common.Input),
		outputs: make(map[uuid.UUID]common.Output),
	}
}

func (r *Ports) AddInput(in common.Input) {
	r.inputs[in.Id()] = in
}

func (r *Ports) RemoveInput(id uuid.UUID) {
	delete(r.inputs, id)
}

func (r *Ports) AddOutput(out common.Output) {
	r.outputs[out.Id()] = out
}

func (r *Ports) RemoveOutput(id uuid.UUID) {
	delete(r.outputs, id)
}
