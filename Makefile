GODEPS := $(shell find . -name '*.go')

ked: $(GODEPS)
	go build cmd/ked/ked.go

all: ked
	go build ./...
	go test ./...

.PHONY: all
