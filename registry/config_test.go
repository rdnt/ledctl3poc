package registry_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ledctl3/registry"
)

func TestParseConfig(t *testing.T) {
	err := registry.ParseConfig([]byte(`
[
  {
    "type": "array",
    "name": "colors",
    "items": {
      "type": "string",
      "name": "color"
    }
  },
  {
    "type": "integer",
    "name": "windowSize",
    "default": 40,
    "minimum": 2,
    "maximum": 1000
  },
  {
    "type": "float",
    "name": "blackPoint",
    "default": 0.2,
    "minimum": 0,
    "maximum": 1
  }
]
`))

	assert.NoError(t, err)
}
