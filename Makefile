.PHONY: build container image setup start test clean

PROGRAM = scandoc
DOCKER_CONTEXT = default
DOCKER_IMAGE = docker.entangle.net/$(PROGRAM):latest
SCAN_DIR = $(realpath scans)
DEVICE_MAJOR = 189

# load variable overrides
-include .env

export DOCKER_CONTEXT
export DOCKER_IMAGE
export SCAN_DIR
export DEVICE_MAJOR

build: out/$(PROGRAM)-linux-amd64

ui/build: $(wildcard ui/package*.json) ui/public ui/src
	# CI=true is used to treat build warnings as errors
	cd ui && CI=true npm run build

out/$(PROGRAM)-linux-amd64: ui/build $(wildcard *.go)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(wildcard *.go)

image:
	docker --context $(DOCKER_CONTEXT) build --tag $(DOCKER_IMAGE) .

setup:
	cd ui && npm $(if $(filter true,$(CI)),clean-install,install)

start:
	scripts/start.sh

container: image
	scripts/container.sh

test:
	# CI=true is used to run tests non-interactively
	go test -v ./...
	cd ui && CI=true npm test

clean:
	rm -rf out scans ui/build
