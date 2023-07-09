package event

type ListCapabilitiesEvent struct {
	Event
}

func (e ListCapabilitiesEvent) Type() Type {
	return ListCapabilities
}
