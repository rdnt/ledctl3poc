package screen

import (
	"ledctl3/pkg/screencapture/dxgi"
	types2 "ledctl3/pkg/screencapture/types"
	"ledctl3/pkg/uuid"
	"ledctl3/source"
	"ledctl3/source/types"
)

type InputRegistry interface {
	AddInput(i source.Input)
	RemoveInput(id uuid.UUID)
}

type InputProvider struct {
	reg  InputRegistry
	repo types2.DisplayRepository
	ids  map[int]uuid.UUID
}

func New(reg InputRegistry) (*InputProvider, error) {
	dr, err := dxgi.New()
	if err != nil {
		return nil, err
	}

	return &InputProvider{
		reg:  reg,
		repo: dr,
		ids:  make(map[int]uuid.UUID),
	}, nil
}

func (p *InputProvider) reset() error {
	displays, err := p.repo.All()
	if err != nil {
		return err
	}

	var ins []*Input
	for _, d := range displays {
		p.ids[d.Id()] = uuid.New()

		ins = append(ins, &Input{
			id:      p.ids[d.Id()],
			events:  make(chan types.UpdateEvent),
			display: d,
			repo:    p.repo,
			outputs: nil,
		})
	}

	return nil
}

func (p *InputProvider) Inputs() ([]*Input, error) {
	displays, err := p.repo.All()
	if err != nil {
		return nil, err
	}

	var ins []*Input
	for _, d := range displays {
		p.ids[d.Id()] = uuid.New()

		ins = append(ins, &Input{
			id:      p.ids[d.Id()],
			events:  make(chan types.UpdateEvent),
			display: d,
			repo:    p.repo,
			outputs: nil,
		})
	}

	return ins, nil
}
