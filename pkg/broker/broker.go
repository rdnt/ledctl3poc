package broker

import (
	"sync"

	"github.com/google/uuid"
)

type Broker[C comparable, E any] struct {
	lock          sync.Mutex
	subscriptions map[C]map[string]func(E)
}

func New[C comparable, E any]() *Broker[C, E] {
	return &Broker[C, E]{
		subscriptions: map[C]map[string]func(E){},
	}
}

func (o *Broker[C, E]) Subscribe(channel C, handler func(e E)) (dispose func()) {
	id := uuid.NewString()

	o.lock.Lock()
	defer o.lock.Unlock()

	if _, ok := o.subscriptions[channel]; !ok {
		o.subscriptions[channel] = map[string]func(E){}
	}

	o.subscriptions[channel][id] = handler

	return func() {
		o.dispose(channel, id)
	}
}

func (o *Broker[C, E]) Publish(channel C, e E) {
	o.lock.Lock()
	defer o.lock.Unlock()

	// TODO: remove simulated network delay
	//time.Sleep(10 * time.Millisecond)

	for ch, subs := range o.subscriptions {
		if ch != channel {
			continue
		}

		for _, h := range subs {
			if h != nil {
				go h(e)
			}
		}
	}
}

func (o *Broker[C, E]) dispose(channel C, id string) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if _, ok := o.subscriptions[channel]; !ok {
		return
	}

	delete(o.subscriptions[channel], id)
}
