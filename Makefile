GODEPS := $(shell find . -name '*.go')
VERSION := $(shell git rev-parse --short HEAD)
BUILDTIME := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GOVARS += -X main.version=$(VERSION)
GOVARS += -X main.buildtime=$(BUILDTIME)
GOFLAGS := -ldflags "$(GOVARS)"

ked: $(GODEPS) Makefile
	go build $(GOFLAGS) cmd/ked/ked.go

all: ked
	go build ./...
	go test ./...
	go vet ./...

install:
	go install $(GOFLAGS) cmd/ked/ked.go

.PHONY: all install
