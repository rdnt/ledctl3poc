package types

import (
	"context"
)

type DisplayRepository interface {
	All() ([]Display, error)
}

type Display interface {
	Id() int
	Width() int
	Height() int
	X() int
	Y() int
	Resolution() string
	String() string
	Close() error
	Capture(ctx context.Context, pollInterval int) (chan []byte, chan error)
	Orientation() Orientation
}

type Orientation int

const (
	Landscape Orientation = iota
	Portrait
	LandscapeFlipped
	PortraitFlipped
)
