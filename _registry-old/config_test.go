package _registry_old_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ledctl3/_registry-old"
)

func TestParseConfig(t *testing.T) {
	err := _registry_old.ParseConfig([]byte(`
{
  "displays": {
    "type": "array",
    "name": "displays",
    "items": {
      "type": "object",
      "name": "display",
      "properties": {
        "width": {
          "name": "width",
          "type": "integer",
          "default": 1920,
          "minimum": 1
        },
        "height": {
          "name": "height",
          "type": "integer",
          "default": 1080,
          "minimum": 1
        },
        "left": {
          "name": "left",
          "type": "integer",
          "default": 0
        },
        "right": {
          "name": "top",
          "type": "integer",
          "default": 0
        },
        "framerate": {
          "name": "framerate",
          "type": "integer",
          "default": 60,
          "minimum": 1
        }
      }
    }
  }
}
`))

	assert.NoError(t, err)
}
