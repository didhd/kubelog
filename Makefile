default: build

GORELEASER := $(shell command -v goreleaser 2> /dev/null)

.PHONY: build release

build:
	env CGO_ENABLED=0 go build -ldflags="-X github.com/didhd/kubelog/pkg.Version=$(shell git describe --tags)" ./cmd/kubelog

install:
	env CGO_ENABLED=0 go install -ldflags="-X github.com/didhd/kubelog/pkg.Version=$(shell git describe --tags)" ./cmd/kubelog

release:
ifndef GORELEASER
	$(error "goreleaser not found (`go get -u -v github.com/goreleaser/goreleaser` to fix)")
endif
	$(GORELEASER) --rm-dist
