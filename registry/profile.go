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
		//return errors.New("profile already enabled") // @@@
	}

	activeOutputIds := r.activeOutputs()

	// cannot have multiple outputs active with different inputs.
	// maybe in the future, input data can be combined to the same
	// output with some transformer function, e.g. one input
	// modifying hue/sat and another modifying brightness.
	for _, linkId := range prof.LinkIds {
		link := r.State.Links[linkId]
		if slices.Contains(activeOutputIds, link.OutputId) {
			//return errors.New("output already in use") // @@@
		}
	}

	prof.Enabled = true

	err := r.sh.SetState(r.State)
	if err != nil {
		return err
	}

	for srcId, src := range r.State.Sources {
		//if !src.Connected {
		//	continue
		//}
		//
		connected := r.State.Nodes[src.NodeId].Connected
		if !connected {
			continue
		}

		err = r.configureSource(srcId)
		if err != nil {
			fmt.Println("error configuring source:", err)
		}
	}

	for _, sink := range r.State.Sinks {
		connected := r.State.Nodes[sink.NodeId].Connected

		if !connected {
			continue
		}

		err := r.writeNode(sink.NodeId, event.ConfigureSink{
			SinkId:     sink.Id,
			SinkConfig: sink.Config,
		})
		if err != nil {
			fmt.Println("error sending event:", err)
			return err
		}
	}

	fmt.Println("profile enabled:", id)

	return nil
}

func (r *Registry) configureSource(srcId uuid.UUID) error {
	src := r.State.Sources[srcId]

	inputCfgs := map[uuid.UUID]event.ConfigureLinks{}

	for _, prof := range r.State.Profiles {
		if !prof.Enabled {
			continue
		}

		for _, linkId := range prof.LinkIds {
			link := r.State.Links[linkId]
			input := r.State.Inputs[link.InputId]
			source := r.State.Sources[input.SourceId]

			if source.Id != srcId {
				continue
			}

			output := r.State.Outputs[link.OutputId]

			cfg, ok := inputCfgs[link.InputId]
			if !ok {
				cfg = event.ConfigureLinks{
					InputId:  input.Id,
					SourceId: input.SourceId,
					Outputs:  nil,
				}
			}

			cfg.Outputs = append(cfg.Outputs, event.ConfigureLinksOutput{
				NodeId:   output.NodeId,
				SinkId:   output.SinkId,
				OutputId: link.OutputId,
				Leds:     output.Leds,
				Config:   link.InputConfig,
			})

			inputCfgs[input.Id] = cfg
		}
	}

	for _, cfg := range inputCfgs {
		err := r.writeNode(src.NodeId, cfg)
		if err != nil {
			fmt.Println("error sending event:", err)
			return err
		}
	}

	return nil
}

//func (r *Registry) enableInputProfile(inputId, profileId uuid.UUID) error {
//	input := r.State.Inputs[inputId]
//	var linkCfgs []event.SetInputActiveOutput
//
//	prof := r.State.Profiles[profileId]
//
//	for _, linkId := range prof.LinkIds {
//		link := r.State.Links[linkId]
//
//		if link.InputId != inputId {
//			continue
//		}
//
//		output := r.State.Outputs[link.OutputId]
//
//		linkCfgs = append(linkCfgs, event.SetInputActiveOutput{
//			NodeId:   output.NodeId,
//			SinkId:   output.SinkId,
//			OutputId: link.OutputId,
//			Leds:     output.Leds,
//			Config:   link.InputConfig,
//		})
//	}
//
//	if len(linkCfgs) == 0 {
//		return nil
//	}
//
//	err := r.writeNode(input.NodeId, event.SetInputActive{
//		Id:      inputId,
//		Outputs: linkCfgs,
//	})
//	if err != nil {
//		fmt.Println("error sending event:", err)
//		return err
//	}
//
//	return nil
//}

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
