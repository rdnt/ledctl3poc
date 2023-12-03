package screen

import "ledctl3/internal/device/common"

type Capturer struct {
}

func New(reg common.InputRegistry) (*Capturer, error) {
	return &Capturer{}, nil
}

func (c *Capturer) Start() {
	return
}
