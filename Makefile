GODEPS := $(shell find . -name '*.go')

ked: $(GODEPS)
	go build cmd/ked/ked.go

all:
	go build ./...
	go test ./...

.PHONY: all
