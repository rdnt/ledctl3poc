#!/bin/bash

#go mod vendor -v

docker run \
    -it \
    --rm \
    --mount type=bind,src="$(pwd)",dst="//app" \
    -w "//app" \
    --platform linux/arm64 \
    --env GOPROXY=direct \
    rpi-ws281x-builder-arm64 \
    bash -c "git config safe.directory '*' && go build -v -o ./build/ledctld ./cmd/node/"
# go build -mod vendor

cp ./build/ledctld ../ledctld/ledctld
#rm -rf ./vendor
