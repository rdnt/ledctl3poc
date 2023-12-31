module ledctl3

go 1.20

require (
	github.com/VividCortex/ewma v1.2.0
	github.com/go-ole/go-ole v1.2.6
	github.com/google/uuid v1.2.0
	github.com/gookit/color v1.5.3
	github.com/grandcat/zeroconf v1.0.1-0.20230119201135-e4f60f8407b1
	github.com/kirides/screencapture v0.0.0-20211031174040-89bc8578d816
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/moutend/go-wca v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/radovskyb/watcher v1.0.7
	github.com/samber/lo v1.38.1
	github.com/sgreben/piecewiselinear v1.1.1
	github.com/stretchr/testify v1.8.4
	github.com/xeipuuv/gojsonschema v1.2.0
	golang.org/x/image v0.6.0
	gonum.org/v1/gonum v0.13.0
	gotest.tools/v3 v3.5.1
)

require (
	github.com/cenkalti/backoff v2.2.1+incompatible // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/miekg/dns v1.1.55 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xo/terminfo v0.0.0-20210125001918-ca9a967f8778 // indirect
	golang.org/x/exp v0.0.0-20230711023510-fffb14384f22 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/net v0.12.0 // indirect
	golang.org/x/sys v0.10.0 // indirect
	golang.org/x/tools v0.11.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/kirides/screencapture v0.0.0-20211031174040-89bc8578d816 => ./pkg/screencapture_kirides
