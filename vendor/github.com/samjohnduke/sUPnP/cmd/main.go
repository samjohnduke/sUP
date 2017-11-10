package main

import (
	"fmt"

	"github.com/samjohnduke/sUPnP"
)

func main() {
	fmt.Println("Hello")

	client, err := sUPnP.Discover()
	if err != nil {
		panic(err)
	}

	a, b, c, err := client.GetRouterStatus()
	if err != nil {
		panic(err)
	}

	fmt.Println(a, b, c)

	pm := &sUPnP.PortMapping{
		RemoteHost:             "203.13.53.45",
		ExternalPort:           7000,
		Protocol:               "TCP",
		InternalClient:         "10.1.1.156",
		InternalPort:           7000,
		PortMappingDescription: "test",
		LeaseDuration:          3600,
	}

	err = client.AddPortMapping(pm)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	mp, err := client.GetPortMappings()
	for _, m := range mp {
		fmt.Printf("%+v\n", m)
	}

	err = client.DeletePortMapping(pm)
	if err != nil {
		panic(err)
	}

	mp, err = client.GetPortMappings()
	for _, m := range mp {
		fmt.Printf("%+v\n", m)
	}

	ip, err := client.GetInternalIP()
	if err != nil {
		panic("unable to connect to router")
	}

	fmt.Println(ip)

	ip, err = client.GetExternalIP()
	if err != nil {
		panic("unable to connect to router")
	}

	fmt.Println(ip)

}
