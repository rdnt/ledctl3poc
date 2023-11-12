package main

import (
	"context"
	"encoding/json"
	"fmt"
	"ledctl3/event"
	"ledctl3/pkg/codec"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver2"
	"ledctl3/pkg/uuid"
	"ledctl3/registry"
	"net"
	"os"
	"time"
)

type registryStore struct {
}

func (r registryStore) SetProfiles(profs map[uuid.UUID]registry.Profile) error {
	b, err := json.MarshalIndent(profs, "", "  ")
	if err != nil {
		return err
	}

	err = os.WriteFile("profiles.json", b, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (r registryStore) Profiles() (map[uuid.UUID]registry.Profile, error) {
	b, err := os.ReadFile("profiles.json")
	if err != nil {
		return nil, err
	}

	var profs map[uuid.UUID]registry.Profile
	err = json.Unmarshal(b, &profs)
	if err != nil {
		return nil, err
	}

	return profs, nil
}

func main() {
	store := registryStore{}

	reg, err := registry.New(store)
	handle(err)

	//port, err := freeport.GetFreePort()
	//if err != nil {
	//	log.Fatal(err)
	//}

	cod := codec.NewGobCodec[event.EventIface](
		[]any{},
		map[string]any{},
		event.AssistedSetupEvent{},
		event.AssistedSetupConfigEvent{},
		event.CapabilitiesEvent{},
		event.ConnectEvent{},
		event.DataEvent{},
		event.ListCapabilitiesEvent{},
		event.SetInputConfigEvent{},
		event.SetSinkActiveEvent{},
		event.SetSourceActiveEvent{},
		event.SetSourceIdleEvent{},
	)

	s := netserver2.New[event.EventIface](-1, cod, func(addr net.Addr, e event.EventIface) {
		reg.ProcessEvent(addr, e)
	})

	go func() {
		for msg := range reg.Messages() {
			err = s.Write(msg.Addr, msg.Payload)
			if err != nil {
				fmt.Print("error sending event: ", err)
			}
		}
	}()

	srv2, err := mdns.NewResolver()
	handle(err)

	devs, err := srv2.Browse(context.Background())
	handle(err)

	go func() {
		for dev := range devs {
			s.Connect(dev.Addr)
			time.Sleep(1 * time.Second)
			err = reg.RegisterDevice(dev.Id, dev.Addr)
			if err != nil {
				fmt.Println(err)
			}

			_, err := reg.AddProfile("profile1", []registry.ProfileSource{
				//{inputdev1a.OutputId(): {outputdev1a.OutputId(), outputdev2b.OutputId()}},
				//{inputdev2b.OutputId(): {outputdev1b.OutputId()}},
				{
					SourceId: uuid.Nil,
					Inputs: []registry.ProfileInput{
						{
							InputId: uuid.Nil,
							Sinks: []registry.ProfileSink{
								{
									SinkId: uuid.Nil,
									Outputs: []registry.ProfileOutput{
										{
											OutputId:      uuid.Nil,
											InputConfigId: uuid.Nil,
										},
									},
								},
							},
						},
					},
				},

				//{inputdev1a.OutputId(): {outputdev2b.OutputId()}}, // audio
				//{inputdev1b.OutputId(): {outputdev2b.OutputId()}}, // video
			})
			handle(err)
		}
	}()

	//time.Sleep(1 * time.Second)

	//err = reg.AssistedSetup(uuid.MustParse("4282186d-dca5-430b-971c-fbe5b9112bfe"))
	//handle(err)

	select {}
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}
