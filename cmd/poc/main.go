package main

import (
	"fmt"
	"github.com/google/uuid"
	"net"
	"os"
	"os/signal"

	"ledctl3/registry"
	"ledctl3/registry/types"
)

func main() {

	reg := registry.New()

	srcAddr := &net.TCPAddr{
		IP:   net.IPv4(192, 168, 1, 10),
		Port: 1234,
	}

	addr := &net.TCPAddr{
		IP:   net.IPv4(192, 168, 1, 11),
		Port: 1234,
	}

	addr2 := &net.TCPAddr{
		IP:   net.IPv4(192, 168, 1, 12),
		Port: 1234,
	}

	addr3 := &net.TCPAddr{
		IP:   net.IPv4(192, 168, 1, 13),
		Port: 1234,
	}

	src := types.NewSource("src", srcAddr)

	d1 := types.NewDevice("d1", addr)
	d1.SetLeds(10)

	d2 := types.NewDevice("d2", addr2)
	d2.SetLeds(20)

	d3 := types.NewDevice("d3", addr3)
	d3.SetLeds(40)

	vdev, err := types.NewVirtualDevice("vdev", d1, d3)
	if err != nil {
		panic(err)
	}

	err = reg.AddDevice(d1)
	if err != nil {
		panic(err)
	}

	err = reg.AddDevice(d2)
	if err != nil {
		panic(err)
	}

	err = reg.AddDevice(d3)
	if err != nil {
		panic(err)
	}

	err = reg.AddDevice(vdev)
	if err != nil {
		panic(err)
	}

	err = reg.AddSource(src)
	if err != nil {
		panic(err)
	}

	devs := reg.Devices()
	for _, dev := range devs {
		fmt.Println(dev)
	}

	prof := reg.AddProfile("ambilight", []map[uuid.UUID]uuid.UUID{
		{src.Id(): vdev.Id()},
	})

	err = reg.SelectProfile(prof.Id, types.StateAmbilight)
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c
}
