package types

type State string

const (
	StateOffline State = "offline"
	StateIdle    State = "idle"
	StateActive  State = "active"
)
