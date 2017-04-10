#!/usr/bin/make -f

build:
	./hack/build.sh linux server provisioner github.com/previousnext/provisioner
	./hack/build.sh linux server status github.com/previousnext/status
	./hack/build.sh linux cli client github.com/previousnext/cli

docker:
	docker build -t previousnext/k8s-aws-efs .

.PHONY: build docker
