.SILENT :

# App name
APPNAME=imgcast

# Go configuration
GOOS?=$(shell go env GOHOSTOS)
GOARCH?=$(shell go env GOHOSTARCH)

# Add exe extension if windows target
is_windows:=$(filter windows,$(GOOS))
EXT:=$(if $(is_windows),".exe","")

# Archive name
ARCHIVE=$(APPNAME)-$(GOOS)-$(GOARCH).tgz

# Executable name
EXECUTABLE=$(APPNAME)$(EXT)

LDFLAGS=-s -w -buildid=

all: build

# Include common Make tasks
root_dir:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))
-include $(root_dir)/.env

.SILENT:

## This help screen
help:
	printf "Available targets:\n\n"
	awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "%-15s %s\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
.PHONY: help

## Clean built files
clean:
	echo ">>> Cleanup..."
	-rm -rf release
.PHONY: clean

## Build executable
build:
	-mkdir -p release
	echo ">>> Building $(EXECUTABLE) for $(GOOS)-$(GOARCH) ..."
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -tags osusergo,netgo -ldflags "$(LDFLAGS)" -o release/$(EXECUTABLE)
.PHONY: build

release/$(EXECUTABLE): build

# Check code style
check-style:
	echo ">>> Checking code style..."
	go vet ./...
	go run honnef.co/go/tools/cmd/staticcheck@latest ./...
.PHONY: check-style

# Check code criticity
check-criticity:
	echo ">>> Checking code criticity..."
	go run github.com/go-critic/go-critic/cmd/gocritic@latest check -enableAll ./...
.PHONY: check-criticity

# Check code security
check-security:
	echo ">>> Checking code security..."
	go run github.com/securego/gosec/v2/cmd/gosec@latest -quiet ./...
.PHONY: check-security

## Code quality checks
checks: check-style check-criticity
.PHONY: checks

## Install executable
install: release/$(EXECUTABLE)
	echo ">>> Installing $(EXECUTABLE) to ${HOME}/.local/bin/$(EXECUTABLE) ..."
	cp release/$(EXECUTABLE) ${HOME}/.local/bin/$(EXECUTABLE)
.PHONY: install

## Create Docker image
image:
	echo ">>> Building Docker image..."
	docker build --rm -t ncarlier/$(APPNAME) .
.PHONY: image

## Create archive
archive: release/$(EXECUTABLE)
	echo ">>> Creating release/$(ARCHIVE) archive..."
	tar czf release/$(ARCHIVE) README.md LICENSE -C release/ $(EXECUTABLE)
	rm release/$(EXECUTABLE)
.PHONY: archive

## Create distribution binaries
distribution:
	GOARCH=amd64 make build archive
	GOARCH=arm64 make build archive
	GOARCH=arm make build archive
	GOOS=darwin make build archive
	GOOS=windows make build archive
.PHONY: distribution
