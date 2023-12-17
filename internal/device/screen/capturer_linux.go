package screen

import (
	"errors"

	"ledctl3/pkg/screencapture/types"
)

func newDisplayRepo(typ string) (types.DisplayRepository, error) {
	switch typ {
	default:
		return nil, errors.New("invalid capturer type")
	}
}
