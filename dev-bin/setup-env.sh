#!/bin/bash

set -e

if [ ! -d "$GOROOT" ]; then
    echo "\$GOROOT of '$GOROOT' does not exist"
    exit 1
fi

pushd "$GOROOT/src"
GOOS=windows GOARCH=amd64 ./make.bash --no-clean
GOOS=linux GOARCH=amd64 ./make.bash --no-clean
GOOS=darwin GOARCH=amd64 ./make.bash --no-clean
popd
