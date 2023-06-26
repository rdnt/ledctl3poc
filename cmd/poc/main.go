package main

import (
	"fmt"
	"net"

	"ledctl3/registry"
	"ledctl3/registry/types"
)

func main() {
	reg := registry.New()

	addr := &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 1234,
	}

	d1 := types.NewDevice(addr)

	vdev, err := types.NewVirtualDevice(d1.Id())
	if err != nil {
		panic(err)
	}

	reg.AddDevice(vdev)

	devs := reg.Devices()

	fmt.Println(devs)
}
