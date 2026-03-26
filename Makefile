VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS  = -X github.com/mshddev/sonacli/internal/cli.Version=$(VERSION)

.PHONY: build test clean

build:
	go build -ldflags "$(LDFLAGS)" -o sonacli .

test:
	go test ./...

clean:
	rm -f sonacli
