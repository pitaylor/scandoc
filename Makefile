.PHONY: build start test clean

PROGRAM = scandoc

SCANNER_DEVICE_MAJOR = 189

build: out/$(PROGRAM)-linux-amd64

out/$(PROGRAM)-linux-amd64: $(wildcard *.go)
	GOOS=linux GOARCH=amd64 go build -o $@ $^

start:
	bash -c "PATH=\"$(shell pwd)/bin:$(PATH)\" go run ."

start_remote: build
	scp out/$(PROGRAM)-linux-amd64 bemo.entangle.net:/tmp/$(PROGRAM).testbuild

	docker --context production run --rm -it \
		-v /dev/bus:/dev/bus:ro \
		-v /dev/serial:/dev/serial:ro \
		-v /mnt/ptaylor/scans:/work/scans \
		-v /tmp/$(PROGRAM).testbuild:/usr/local/bin/scandoc \
		-p 8090:8090 \
		--cap-add SYS_PTRACE \
		--device-cgroup-rule "c $(SCANNER_DEVICE_MAJOR):* rwm" \
		homelab_scanner \
		scandoc -dir /work/scans

test:
	go test -v ./...

clean:
	rm -rf out scans
