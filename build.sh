#!/bin/sh

set -ex

GOBIN=$PWD/bin go install -v ./cmd/compile_assets
$PWD/bin/compile_assets
GOBIN=$PWD/bin go install -v ./cmd/...
