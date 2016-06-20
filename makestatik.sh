#!/bin/bash
set -xeu -o pipefail

export PATH="$GOPATH/bin:$PATH"
which statik || go get github.com/rakyll/statik
rm -rf statik
statik -src=assets
