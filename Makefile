VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

.PHONY: build test lint clean install golden

build:
	go build $(LDFLAGS) -o dist/shedoc ./cmd/shedoc

test:
	go test ./...

lint:
	golangci-lint run

clean:
	rm -rf dist/

install:
	go install $(LDFLAGS) ./cmd/shedoc

golden:
	go test ./... -update
