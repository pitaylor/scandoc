.PHONY: all build container image start test clean

PROGRAM = scandoc
DOCKER_CONTEXT = default
DOCKER_IMAGE = docker.entangle.net/$(PROGRAM):latest
SCAN_DIR = $(realpath scans)
DEVICE_MAJOR = 189

# load variable overrides
-include .env

build: out/$(PROGRAM)-linux-amd64

out/$(PROGRAM)-linux-amd64: $(wildcard *.go)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $^

image:
	docker --context $(DOCKER_CONTEXT) build --tag $(DOCKER_IMAGE) .

all: build image

start:
	bash -c "PATH=\"$(realpath bin):$(PATH)\" go run ."

container: image
	docker --context $(DOCKER_CONTEXT) run --rm -it \
		-v /dev/bus:/dev/bus:ro \
		-v /dev/serial:/dev/serial:ro \
		-v "$(SCAN_DIR):/work/scans" \
		-p 8090:8090 \
		--cap-add SYS_PTRACE \
		--device-cgroup-rule "c $(DEVICE_MAJOR):* rwm" \
		$(DOCKER_IMAGE) \
		scandoc -dir /work/scans

test:
	go test -v ./...

clean:
	rm -rf out scans
