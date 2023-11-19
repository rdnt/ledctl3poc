package fsbroker

import (
	"encoding/gob"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/radovskyb/watcher"

	"ledctl3/event"
	"ledctl3/pkg/uuid"
)

func init() {
	//gob.Register(&event.AssistedSetup{})
	//gob.Register(&event.AssistedSetupConfig{})
	//gob.Register(&event.Capabilities{})
	//gob.Register(&event.Connect{})
	//gob.Register(&event.Data{})
	//gob.Register(&event.ListCapabilities{})
	//gob.Register(&event.SetInputConfig{})
	//gob.Register(&event.SetSinkActive{})
	//gob.Register(&event.SetSourceActive{})
	//gob.Register(&event.SetSourceIdle{})

	gob.Register(event.AssistedSetup{})
	gob.Register(event.AssistedSetupConfig{})
	gob.Register(event.Capabilities{})
	gob.Register(map[string]any{})
	gob.Register([]any{})
	gob.Register(event.Connect{})
	gob.Register(event.Data{})
	gob.Register(event.ListCapabilities{})
	gob.Register(event.SetInputConfig{})
	gob.Register(event.SetSinkActive{})
	gob.Register(event.SetSourceActive{})
	gob.Register(event.SetSourceIdle{})
}

type Broker[E any] struct {
	watcher       *watcher.Watcher
	lock          sync.Mutex
	subscriptions map[string]map[uuid.UUID]func(E)
}

func New[E any]() *Broker[E] {
	return &Broker[E]{
		subscriptions: map[string]map[uuid.UUID]func(E){},
		watcher:       watcher.New(),
	}
}

func (o *Broker[E]) Subscribe(ch uuid.UUID, handler func(e E)) (dispose func()) {
	fmt.Println("SUBSCRIBE", ch)
	id := uuid.New()

	o.lock.Lock()
	defer o.lock.Unlock()

	channel := ch.String()

	if _, ok := o.subscriptions[channel]; !ok {
		o.subscriptions[channel] = map[uuid.UUID]func(E){}
		f, err := os.Create("evts/" + channel)
		if err != nil {
			fmt.Println(err)
		}
		f.Close()
		//
		//enc := gob.NewEncoder(f)
		//err = enc.Encode(nil)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}

		err = o.Watch("evts/" + channel)
		if err != nil {
			fmt.Println(err)
		}
	}

	o.subscriptions[channel][id] = handler

	return func() {
		o.dispose(channel, id)
	}
}

func (o *Broker[E]) Publish(channel uuid.UUID, e E) {
	fmt.Println("PUBLISH", channel, e)

	o.lock.Lock()
	defer o.lock.Unlock()

	//var skip bool
	f, err := os.OpenFile("evts/"+channel.String(), os.O_RDWR|os.O_TRUNC, 0644)
	if errors.Is(err, os.ErrNotExist) {
		panic(err)
		//f, err = os.OpenFile("evts/"+channel.String(), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		//if err != nil {
		//	fmt.Println(err)
		//	return
		//}
		//skip = true
	} else if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	//if !skip {
	//	dec := gob.NewDecoder(f)
	//	err = dec.Decode(&evts)
	//	if err != nil {
	//		fmt.Println(err)
	//		return
	//	}
	//}

	enc := gob.NewEncoder(f)
	err = enc.Encode(&e)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func (o *Broker[E]) dispose(channel string, id uuid.UUID) {
	o.lock.Lock()
	defer o.lock.Unlock()

	if _, ok := o.subscriptions[channel]; !ok {
		return
	}

	delete(o.subscriptions[channel], id)
}

func (w *Broker[E]) Watch(path string) error {
	err := w.watcher.Add(path)
	if err != nil {
		return err
	}
	return nil
}

func (w *Broker[E]) Start() {
	fmt.Println("starting watcher")
	go func() {
		for {
			select {
			case evt := <-w.watcher.Event:
				fmt.Println("READ")

				f, err := os.Open(evt.Path)
				if err != nil {
					fmt.Println(err)
					continue
				}

				//b, err := io.ReadAll(f)
				//if err != nil {
				//	fmt.Println(err)
				//	continue
				//}
				//fmt.Println(string(b))
				//f.Seek(0, 0)

				dec := gob.NewDecoder(f)

				var e E
				err = dec.Decode(&e)
				if err != nil {
					f.Close()
					fmt.Println(err)
					continue
				}
				f.Close()

				fmt.Println("RECEIVED EVENT", e)

				for ch, subs := range w.subscriptions {
					if ch != evt.Name() {
						continue
					}

					for _, h := range subs {
						if h != nil {
							go func(h func(E)) {
								// TODO: remove simulated network delay
								//time.Sleep(10 * time.Millisecond)
								h(e)
							}(h)
						}
					}
				}

			case err := <-w.watcher.Error:
				fmt.Println(err)
				continue
			case <-w.watcher.Closed:
				return
			}
		}
	}()

	go func() {
		_ = w.watcher.Start(1 * time.Millisecond)
	}()
}

func (o *Broker[E]) Stop() {
	o.watcher.Close()
}
