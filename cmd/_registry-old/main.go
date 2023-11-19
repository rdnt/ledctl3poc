package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/samber/lo"

	"ledctl3/_registry-old"
	"ledctl3/event"
	"ledctl3/pkg/mdns"
	"ledctl3/pkg/netserver2"
	"ledctl3/pkg/uuid"
)

type registryStore struct {
}

//func (r registryStore) Sources() (map[uuid.UUID]*source.Source, error) {
//	b, err := os.ReadFile("sources.json")
//	if err != nil {
//		return nil, err
//	}
//
//	var srcs map[uuid.UUID]*source.Source
//	err = json.Unmarshal(b, &srcs)
//	if err != nil {
//		return nil, err
//	}
//
//	return srcs, nil
//}
//
//func (r registryStore) SetSources(src map[uuid.UUID]*source.Source) error {
//	b, err := json.MarshalIndent(src, "", "  ")
//	if err != nil {
//		return err
//	}
//
//	err = os.WriteFile("sources.json", b, 0644)
//	if err != nil {
//		return err
//	}
//
//	return nil
//}

func (r registryStore) SetProfiles(profs map[uuid.UUID]_registry_old.Profile) error {
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

func (r registryStore) Profiles() (map[uuid.UUID]_registry_old.Profile, error) {
	b, err := os.ReadFile("profiles.json")
	if err != nil {
		return nil, err
	}

	var profs map[uuid.UUID]_registry_old.Profile
	err = json.Unmarshal(b, &profs)
	if err != nil {
		return nil, err
	}

	return profs, nil
}

func main() {
	store := registryStore{}

	reg, err := _registry_old.New(store)
	handle(err)

	//port, err := freeport.GetFreePort()
	//if err != nil {
	//	log.Fatal(err)
	//}

	s := netserver2.New[event.EventIface](1337, event.Codec, nil, func(addr net.Addr, e event.EventIface) {
		reg.ProcessEvent(addr, e)
	})

	err = s.Start()
	handle(err)

	go func() {
		for msg := range reg.Messages() {
			err = s.Write(msg.Addr, msg.Payload)
			if err != nil {
				fmt.Print("error sending event: ", err)
			}
		}
	}()

	mdnsServer, err := mdns.NewServer("registry", 1337)
	handle(err)

	err = mdnsServer.Start()
	handle(err)

	time.Sleep(1 * time.Second)

	//_, err = reg.AddProfile("profile1", []registry.ProfileSource{
	//	//{inputdev1a.Id(): {outputdev1a.Id(), outputdev2b.Id()}},
	//	//{inputdev2b.Id(): {outputdev1b.Id()}},
	//	{
	//		Id: uuid.Nil,
	//		Inputs: []registry.ProfileInput{
	//			{
	//				Id: uuid.Nil,
	//				Sinks: []registry.ProfileSink{
	//					{
	//						Id: uuid.Nil,
	//						Outputs: []registry.ProfileOutput{
	//							{
	//								Id:      uuid.Nil,
	//								InputConfigId: uuid.Nil,
	//							},
	//						},
	//					},
	//				},
	//			},
	//		},
	//	},
	//
	//	//{inputdev1a.Id(): {outputdev2b.Id()}}, // audio
	//	//{inputdev1b.Id(): {outputdev2b.Id()}}, // video
	//})
	//handle(err)

	//err = reg.AssistedSetup(uuid.MustParse("4282186d-dca5-430b-971c-fbe5b9112bfe"))
	//handle(err)

	time.Sleep(1 * time.Second)

	err = reg.AssistedSetup(uuid.MustParse("72c04693-fe20-433d-8dd0-1a8892960e95"))
	handle(err)

	//err = reg.AssistedSetup(uuid.MustParse("59d757d3-51e2-4baf-b008-894f26d3f689"))
	//handle(err)

	time.Sleep(3 * time.Second)

	cfgs := lo.Values(reg.InputConfigs(uuid.MustParse("72c04693-fe20-433d-8dd0-1a8892960e95")))

	if len(cfgs) == 0 {
		panic("no config")
	}

	cfgs[0].Cfg["reverse"] = true

	reg.UpdateInputConfig(uuid.MustParse("72c04693-fe20-433d-8dd0-1a8892960e95"), cfgs[0].Id, "custom", cfgs[0].Cfg)

	// TODO: input configs need to be persisted

	err = reg.SelectProfile(uuid.MustParse("974f0075-59a2-4421-8afb-b7ef61b6a3e5"), true)
	handle(err)

	fmt.Println("idle")
	select {}
}

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

//var t = event.SetSourceActive{
//	Event:     event.Event{Type: "setActive"},
//	SessionId: "c63c1eeb-bdf7-4757-9b12-ba52ddb8c3d3",
//	Inputs: []event.SetSourceActiveInput{{
//		Id: "60ff01a7-5507-48f7-8ae5-dd6c8609f98b",
//		Sinks: []event.SetSourceActiveSink{{
//			Id: "d17c94aa-2fb1-4fb5-b315-f22113e8d165",
//			Outputs: []event.SetSourceActiveOutput{{
//				Id:     "30dc1242-2f66-4fb9-8db0-d8f29beca51c",
//				Config: map[string]interface{}(nil),
//				Leds:   20,
//			}},
//		}},
//	}, {
//		Id: "59d757d3-51e2-4baf-b008-894f26d3f689",
//		Sinks: []event.SetSourceActiveSink{{
//			Id: "d17c94aa-2fb1-4fb5-b315-f22113e8d165",
//			Outputs: []event.SetSourceActiveOutput{{
//				Id:     "c715765b-29a9-42e3-aec6-f590978fb1dd",
//				Config: map[string]interface{}(nil),
//				Leds:   120,
//			}},
//		}},
//	}},
//}
