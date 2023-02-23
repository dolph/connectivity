#!/bin/bash
set -e

# Vet
go vet

# Build
GIT_COMMIT="$(git rev-parse --short $(git rev-list -1 HEAD))"
GIT_TAG="$(git tag --points-at HEAD)"
GO_VERSION="$(go version | cut -d' ' -f3)"
BUILD_DATE="$(date --utc)"
BUILD_ARCH="$(arch)"
if [ "$(git status -s | wc -l)" -eq "0" ]; then
    BUILD_TAINTED=false
else
    BUILD_TAINTED=true
fi
go build \
    -v \
    -ldflags " \
        -X 'main.GitTag=$GIT_TAG' \
        -X 'main.GitCommit=$GIT_COMMIT' \
        -X 'main.GoVersion=$GO_VERSION' \
        -X 'main.BuildTimestamp=$BUILD_DATE' \
        -X 'main.BuildOS=$OSTYPE' \
        -X 'main.BuildArch=$BUILD_ARCH' \
        -X 'main.BuildTainted=$BUILD_TAINTED'" \
    ./...

# Test
go test -cover -v ./...
