package screen

import (
	"errors"

	"ledctl3/pkg/screencapture/dxgi"
	"ledctl3/pkg/screencapture/types"
)

func newDisplayRepo(typ string) (types.DisplayRepository, error) {
	switch typ {
	case "dxgi":
		return dxgi.New()
	default:
		return nil, errors.New("invalid capturer type")
	}
}
