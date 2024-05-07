package registry

import (
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"

	"ledctl3/node/event"
	"ledctl3/registry/types"
)

func (r *Registry) CreateProfile(id uuid.UUID, name string, links []*types.Link) error {
	var linkIds []uuid.UUID
	for _, l := range links {
		l.Id = uuid.New()
		l.ProfileId = id
		linkIds = append(linkIds, l.Id)

		r.State.Links[l.Id] = l
	}

	prof := &types.Profile{
		Id:      id,
		Name:    name,
		LinkIds: linkIds,
	}

	r.State.Profiles[prof.Id] = prof

	return r.sh.SetState(r.State)
}

func (r *Registry) EnableProfile(id uuid.UUID) error {
	prof, ok := r.State.Profiles[id]
	if !ok {
		return errors.New("profile not found")
	}

	if prof.Enabled {
		return errors.New("profile already enabled")
	}

	activeOutputIds := r.activeOutputs()

	// cannot have multiple outputs active with different inputs.
	// maybe in the future, input data can be combined to the same
	// output with some transformer function, e.g. one input
	// modifying hue/sat and another modifying brightness.
	for _, linkId := range prof.LinkIds {
		link := r.State.Links[linkId]
		if slices.Contains(activeOutputIds, link.OutputId) {
			return errors.New("output already in use")
		}
	}

	prof.Enabled = true

	err := r.sh.SetState(r.State)
	if err != nil {
		return err
	}

	fmt.Println("profile enabled:", id)

	for _, linkId := range prof.LinkIds {
		link := r.State.Links[linkId]
		srcNodeId := r.State.Inputs[link.InputId].NodeId
		output := r.State.Outputs[link.OutputId]

		err = r.writeNode(srcNodeId, event.SetInputActive{
			Id: link.InputId,
			Outputs: []event.SetInputActiveOutput{
				{
					OutputId: link.OutputId,
					Leds:     output.Leds,
					Config:   link.InputConfig,
				},
			},
		})
		if err != nil {
			fmt.Println("error sending event:", err)
			continue
		}
	}

	return nil
}

func (r *Registry) activeOutputs() []uuid.UUID {
	var outputIds []uuid.UUID

	for _, prof := range r.State.Profiles {
		if !prof.Enabled {
			continue
		}

		for _, linkId := range prof.LinkIds {
			link := r.State.Links[linkId]
			outputIds = append(outputIds, link.OutputId)
		}
	}

	return outputIds
}
