package registry_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ledctl3/registry"
)

func TestParseConfig(t *testing.T) {
	err := registry.ParseConfig([]byte(`
{
  "colors": {
    "type": "array",
    "name": "Colors",
    "items": {
      "type": "string",
      "name": "Color"
    }
  },
  "windowSize": {
    "type": "integer",
    "name": "Window size",
    "default": 40,
    "minimum": 2,
    "maximum": 1000
  },
  "blackPoint": {
    "type": "float",
    "name": "Black point",
    "default": 0.2,
    "minimum": 0,
    "maximum": 1
  }
}
`))

	assert.NoError(t, err)
}
