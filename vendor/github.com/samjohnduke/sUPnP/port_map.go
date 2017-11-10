package sUPnP

// PortMapping A UPnP port map as described by the interface
type PortMapping struct {
	RemoteHost             string
	ExternalPort           uint16
	Protocol               string
	InternalPort           uint16
	InternalClient         string
	Enabled                bool
	PortMappingDescription string
	LeaseDuration          uint32
}
