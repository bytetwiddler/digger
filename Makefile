.PHONY: all build test clean run build-windows build-linux build-darwin build-all build-service-manager

VERSION := $(shell git describe --tags --always --dirty)
BUILD := $(shell git rev-parse --short HEAD)
DATE := $(shell powershell -Command "Get-Date -Format yyyy-MM-dd")
LDARGS := -X main.version=$(VERSION) -X main.build=$(BUILD) -X main.date=$(DATE)
OUTPUT_DIR := build

all: build

build:
	go build -ldflags "$(LDARGS)" -o digger -v ./cmd/digger

build-windows:
	if not exist $(OUTPUT_DIR) mkdir $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)\digger-windows-amd64.exe -v .\cmd\digger
	copy /Y config.yaml $(OUTPUT_DIR)

build-service-manager:
	if not exist $(OUTPUT_DIR) mkdir $(OUTPUT_DIR)
	go build -ldflags "$(LDARGS)" -o $(OUTPUT_DIR)\service_manager.exe -v .\cmd\service_manager
	copy /Y config.yaml $(OUTPUT_DIR)

build-all: build-windows build-service-manager

test:
	go test -v ./... -cover

clean:
	if exist $(OUTPUT_DIR) rd /s /q $(OUTPUT_DIR)
	if exist digger.exe del /q digger.exe
	if exist digger.log del /q digger.log

run: build
	.\digger.exe