package registry

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/google/uuid"

	"ledctl3/node/event"
	"ledctl3/registry/types"
)

func (r *Registry) write(c Conn, v event.Event) error {
	codec := event.JSONCodec{}

	if _, ok := v.(event.Data); !ok {
		fmt.Printf("%s: send %s\n", c.Addr(), reflect.TypeOf(v).Name())
	}

	b, err := codec.MarshalBinary(v)
	if err != nil {
		return err
	}

	return c.Write(b)
}

func (r *Registry) writeNode(nodeId uuid.UUID, v event.Event) error {
	addr, ok := r.NodeAddr[nodeId]
	if !ok {
		return errors.New(fmt.Sprint("node not found ", nodeId))
	}

	c, ok := r.Conns[addr]
	if !ok {
		return errors.New("node not connected")
	}

	return r.write(c, v)
}

func (r *Registry) activeNodeOutputs(nodeId uuid.UUID) []uuid.UUID {
	var outputIds []uuid.UUID

	for _, prof := range r.State.Profiles {
		for _, linkId := range prof.LinkIds {
			link := r.State.Links[linkId]
			output := r.State.Outputs[link.OutputId]

			if output.NodeId != nodeId {
				continue
			}

			outputIds = append(outputIds, link.OutputId)
		}
	}

	return outputIds
}

func (r *Registry) activeInputLinks(id uuid.UUID) []*types.Link {
	var cfgs []*types.Link

	for _, prof := range r.State.Profiles {
		if !prof.Enabled {
			continue
		}

		for _, linkId := range prof.LinkIds {
			link := r.State.Links[linkId]

			if link.InputId != id {
				continue
			}

			cfgs = append(cfgs, link)
		}
	}

	return cfgs
}

func activateInput() {
	//cfgs := r.activeInputLinks(e.OutputId)
	//if len(cfgs) == 0 {
	//	return nil
	//}
	//
	//var evtOutCfgs []event.SetInputActiveOutput
	//for _, cfg := range cfgs {
	//	dev := r.State.Nodes[r.outputNodeId(cfg.OutputId)]
	//
	//	// TODO: validate output exists
	//
	//	evtOutCfgs = append(evtOutCfgs, event.SetInputActiveOutput{
	//		OutputId: cfg.OutputId,
	//		SinkId:   dev.Outputs[cfg.OutputId].DriverId,
	//		Leds:     dev.Outputs[cfg.OutputId].Leds,
	//		Config:   cfg.Config,
	//	})
	//}
	//
	//err := r.write(c, event.SetInputActive{
	//	OutputId:      e.OutputId,
	//	Config:  e.Config,
	//	Outputs: evtOutCfgs,
	//})
	//if err != nil {
	//	return err
	//}
	//
	//fmt.Println("sent SetInputActive to", c.Addr())
}
