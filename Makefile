version := $(shell git describe --tags)
revision := $(shell git rev-parse HEAD)
release := $(shell git describe --tags | cut -d"-" -f 1,2)
build_date := $(shell date -u +"%Y-%m-%dT%H:%M:%S+00:00")
application := $(shell basename `pwd`)

GO_LDFLAGS := "-X github.com/jnovack/go-version.Application=${application} -X github.com/jnovack/go-version.Version=${version} -X github.com/jnovack/go-version.Revision=${revision} -X github.com/jnovack/go-version.BuildDate=${build_date}"

all: build

.PHONY: install
install:
	cp cloudkey-screen.service /lib/systemd/system/cloudkey-screen.service
	cp cloudkey-screen /usr/local/bin/cloudkey-screen

.PHONY: build
build:
	GOOS=linux GOARCH=arm go build -ldflags $(GO_LDFLAGS) cloudkey-screen.go
