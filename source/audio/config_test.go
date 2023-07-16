package audio_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"ledctl3/source/audio"
)

func TestConfig(t *testing.T) {
	a := &audio.AudioCapture{}
	err := a.ApplyConfig([]byte(`
{
    "selectedProfile": "a7d5fbcf-698a-406e-ae8c-250e7f14bb79",
      "profiles": [
        {
          "id": "a7d5fbcf-698a-406e-ae8c-250e7f14bb79",
          "name": "cyberpunk",
          "colors": [
            "#4a1524",
            "#065394",
            "#00b585",
            "#d600a4",
            "#ff004c"
          ],
          "windowSize": 2,
          "blackPoint": 0.2
        },
        {
          "id": "84d30586-d643-412e-ab70-20a80600fe75",
          "name": "yellow",
          "colors": [
            "#FFB87A",
            "#FFD77A"
          ],
          "windowSize": 2,
          "blackPoint": 0.2
        },
        {
          "id": "d7cfd73d-fb7c-4d9f-939b-e5a7416c35e0",
          "name": "warm",
          "colors": [
            "#FFB87A",
            "#FFB87A",
            "#FFB87A",
            "#ff8658"
          ],
          "windowSize": 2,
          "blackPoint": 0.2
        },
        {
          "id": "b8817db1-e74a-4563-bd23-7cca55960a14",
          "name": "simple",
          "colors": [
            "#ffdc7b",
            "#ffdc7b"
          ],
          "windowSize": 2,
          "blackPoint": 0.2
        },
        {
          "id": "0475d231-7b13-4ece-8ad0-eadb891cdc3d",
          "name": "rgb",
          "colors": [
            "#ff0000",
            "#00ff00",
            "#0000ff",
            "#0000ff"
          ],
          "windowSize": 2,
          "blackPoint": 0.2
        },
        {
          "id": "3b4f38bc-4c4f-4175-b99e-d6b036176bf7",
          "name": "dunno",
          "colors": [
            "#c7839c",
            "#ede5ce",
            "#016d9c",
            "#004972",
            "#c7839c",
            "#b4fbfc"
          ],
          "windowSize": 2,
          "blackPoint": 0.2
        }
      ]
}
`))
	assert.NoError(t, err)
}
