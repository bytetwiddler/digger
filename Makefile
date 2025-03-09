.PHONY: all build test clean run build-windows build-linux build-darwin build-all

VERSION := $(shell git describe --tags --always --dirty)
BUILD := $(shell git rev-parse --short HEAD)
DATE := $(shell date +%Y-%m-%d)
LDARGS := -X main.version=$(VERSION) -X main.build=$(BUILD) -X main.date=$(DATE)
OUTPUT_DIR := build

# Create the output directory if it doesn't exist
$(shell mkdir -p $(OUTPUT_DIR))

all: build

build:
	go build -ldflags "$(LDARGS)" -o digger -v ./cmd/digger

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDARGS)" -o $(OUTPUT_DIR)/digger-windows-amd64.exe -v ./cmd/digger

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDARGS)" -o $(OUTPUT_DIR)/digger-linux-amd64 -v ./cmd/digger

build-darwin:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDARGS)" -o $(OUTPUT_DIR)/digger-darwin-arm64 -v ./cmd/digger

build-all: build-windows build-linux build-darwin

test:
	go test -v ./...

clean:
	rm -f digger $(OUTPUT_DIR)/digger-windows-amd64.exe $(OUTPUT_DIR)/digger-linux-amd64 $(OUTPUT_DIR)/digger-darwin-arm64

run: build
	./digger