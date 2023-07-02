package event

type SetIdleEvent struct {
	Event Type `json:"event"`
}

func (e SetIdleEvent) Type() Type {
	return SetIdle
}
