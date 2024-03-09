package registry

import (
	"context"
	"errors"
	"fmt"

	"ledctl3/node/event"
	"ledctl3/pkg/uuid"
	"ledctl3/registry/req"
)

func (r *Registry) HandleRequest(addr string, e req.Request) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	//fmt.Println("HandleConnection")

	var err error
	switch e := e.(type) {

	case req.SetSourceConfig:
		err = r.handleSetSourceConfig(addr, e)

	default:
		fmt.Printf("unknown command %#v\n", e)
	}

	if err != nil {
		return err
	}

	//fmt.Println("Saving State", fmt.Sprintf("%#v", *r.State))
	err = r.sh.SetState(*r.State)
	if err != nil {
		fmt.Println("error writing State", err)
	}

	//fmt.Println("HandleConnection done")
	return nil
}

func (r *Registry) handleSetSourceConfig(addr string, e req.SetSourceConfig) error {
	fmt.Printf("%s: recv SetSourceConfig\n", addr)

	var node *Node
	for _, nod := range r.State.Nodes {
		_, ok := nod.Sources[e.SourceId]
		if ok {
			node = nod
			break
		}
	}

	if node == nil {
		return errors.New("unknown source")
	}

	_, ok := r.connsAddr[node.Id]
	if !ok {
		return errors.New("node disconnected")
	}

	source, ok := node.Sources[e.SourceId]
	if !ok {
		return errors.New("unknown source")
	}

	return r.SetSourceConfig(node.Id, source.Id, e.Config)
}

func (r *Registry) SetSourceConfig(nodeId, sourceId uuid.UUID, cfg []byte) error {
	r.State.Nodes[nodeId].Sources[sourceId].Config = cfg

	err := r.sh.SetState(*r.State)
	if err != nil {
		return err
	}

	err = r.req(r.connsAddr[nodeId], event.SetSourceConfig{
		SourceId: sourceId,
		Config:   cfg,
	})
	if err != nil {
		fmt.Println("error sending event:", err)
		return err
	}

	return nil
}

func (r *Registry) onResourceEvent(ctx context.Context, id uuid.UUID, h func(string, any) error) {
	for {
		select {
		case <-ctx.Done():

		}
	}
	//r.handlers[id] = append(r.handlers[id], h)
}
