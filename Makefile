out_dir := build/out/bin

REPO = "github.com/zlabjp/openstack-iid-dynamic-json"
VERSION ?= git-$(shell git rev-parse --short HEAD)

uname := $(shell uname -s)
ifeq (${uname},Linux)
	OS=linux
endif
ifeq (${uname},Darwin)
	OS=darwin
endif

build-linux: OS=linux
build-linux: build

build-darwin: OS=darwin
build-darwin: build

build: clean
	cd cmd && GO111MODULE=on GOOS=$(OS) GOARCH=amd64 go build -ldflags "-w -X main.repo=$(REPO) -X main.version=$(VERSION)" -o ../$(out_dir)/openstack-iid-dynamic-json  -i

test:
	GO111MODULE=on go test -race ./cmd/... ./pkg/...

clean:
	go clean
	-rm -rf build/out

.PHONY: all build test clean
