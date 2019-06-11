#!/usr/bin/make -f

export CGO_ENABLED=0

PROJECT=github.com/previousnext/k8s-aws-efs

# Builds the project
build:
	gox -os='linux darwin' -arch='amd64' -output='bin/k8s-aws-efs_{{.OS}}_{{.Arch}}' -ldflags='-extldflags "-static"' $(PROJECT)

# Builds tools associated with the project.
tools:
	gox -os='linux darwin' -arch='amd64' -output='bin/reaper-mount.nfs_{{.OS}}_{{.Arch}}' -ldflags='-extldflags "-static"' $(PROJECT)/tools/reaper-mount.nfs

# Run all lint checking with exit codes for CI
lint:
	golint -set_exit_status `go list ./... | grep -v /vendor/`

# Run tests with coverage reporting
test:
	go test -cover ./...

IMAGE=previousnext/k8s-aws-efs
VERSION=$(shell git describe --tags --always)

# Releases the project Docker Hub
release:
	docker build -t ${IMAGE}:${VERSION} .
	docker push ${IMAGE}:${VERSION}

.PHONY: build lint test release tools
