.PHONY: build start test clean

PROGRAM = docscan

build: out/$(PROGRAM)-linux-amd64

out/$(PROGRAM)-linux-amd64: $(wildcard *.go)
	GOOS=linux GOARCH=amd64 go build -o $@ $^

start:
	bash -c "PATH=\"$(shell pwd):$(PATH)\" go run ."

test:
	go test -v ./...

clean:
	rm -rf out scans
