#!/bin/bash

DIST=$1
TYPE=$2
NAME=$3
MAIN=$4

export GOPATH=$(pwd)/vendor:$(pwd)
export GOOS=${DIST}
export GOARCH=amd64
export CGO_ENABLED=0

DIR=$(pwd)/bin/${GOOS}/${TYPE}

echo "Building: ${GOOS}/${GOARCH}/${TYPE}/${NAME}"

mkdir -p $DIR
go build -a -ldflags '-extldflags "-static"' -o ${DIR}/${NAME} $MAIN
