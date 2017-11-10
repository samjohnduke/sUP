package sUPnP

import (
	"errors"
	"net"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway1"
)

// Client is the container client implmented by GoUPnP
type Client interface {
	GetExternalIPAddress() (string, error)
	AddPortMapping(string, uint16, string, uint16, string, bool, string, uint32) error
	DeletePortMapping(string, uint16, string) error
	GetServiceClient() *goupnp.ServiceClient
	GetGenericPortMappingEntry(uint16) (string, uint16, string, uint16, string, bool, string, uint32, error)
	GetStatusInfo() (string, string, uint32, error)
}

// IGD - The struct that interfaces with the Internet Gateway Device manager instance
type IGD struct {
	c Client
}

//GetPortMappings returns all available port mappings on the router for UPnP
func (igd *IGD) GetPortMappings() ([]*PortMapping, error) {

	var pms = []*PortMapping{}
	var index uint16
	for {
		host, ePort, protocol, iPort, iClient, enabled, description, duration, err := igd.c.GetGenericPortMappingEntry(index)

		if err != nil {
			break
		}

		pm := &PortMapping{
			RemoteHost:             host,
			ExternalPort:           ePort,
			Protocol:               protocol,
			InternalPort:           iPort,
			InternalClient:         iClient,
			Enabled:                enabled,
			PortMappingDescription: description,
			LeaseDuration:          duration,
		}

		pms = append(pms, pm)
		index = index + 1
	}

	return pms, nil
}

//AddPortMapping creates a new port mapping on the router
func (igd *IGD) AddPortMapping(pm *PortMapping) error {
	return igd.c.AddPortMapping(
		pm.RemoteHost,
		pm.ExternalPort,
		pm.Protocol,
		pm.InternalPort,
		pm.InternalClient,
		pm.Enabled,
		pm.PortMappingDescription,
		pm.LeaseDuration,
	)
}

//DeletePortMapping removes a port mapping on the router
func (igd *IGD) DeletePortMapping(pm *PortMapping) error {
	return igd.c.DeletePortMapping(
		pm.RemoteHost,
		pm.ExternalPort,
		pm.Protocol,
	)
}

//GetRouterStatus tells us some information about the router
func (igd *IGD) GetRouterStatus() (string, string, uint32, error) {
	return igd.c.GetStatusInfo()
}

//GetInternalIP returns the callers IP address that they received from the router
func (igd *IGD) GetInternalIP() (string, error) {
	host, _, _ := net.SplitHostPort(igd.c.GetServiceClient().RootDevice.URLBase.Host)
	devIP := net.ParseIP(host)
	if devIP == nil {
		return "", errors.New("could not determine router's internal IP")
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			if x, ok := addr.(*net.IPNet); ok && x.Contains(devIP) {
				return x.IP.String(), nil
			}
		}
	}

	return "", errors.New("could not determine internal IP")
}

//GetExternalIP returns the ip address that an external machine would connect to in order to
//connect via tcp or udp
func (igd *IGD) GetExternalIP() (string, error) {
	return igd.c.GetExternalIPAddress()
}

//Discover searches for a router on the network
func Discover() (*IGD, error) {

	pppclients, _, _ := internetgateway1.NewWANPPPConnection1Clients()
	if len(pppclients) > 0 {
		return &IGD{pppclients[0]}, nil
	}

	ipclients, _, _ := internetgateway1.NewWANIPConnection1Clients()
	if len(ipclients) > 0 {
		return &IGD{ipclients[0]}, nil
	}

	return nil, errors.New("unable to find a uPnP enabled router")
}
