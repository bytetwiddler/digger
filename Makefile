.PHONY: all build test clean run build-windows build-linux build-darwin build-all

VERSION := $(shell git describe --tags --always --dirty)
BUILD := $(shell git rev-parse --short HEAD)
DATE := $(shell powershell -Command "Get-Date -Format yyyy-MM-dd")
LDARGS := -X main.version=$(VERSION) -X main.build=$(BUILD) -X main.date=$(DATE)
OUTPUT_DIR := build

# Change 'all' to build everything
all: build-all

build:
	go build -ldflags "$(LDARGS)" -o digger -v ./cmd/digger

# Windows builds include both digger and service manager
build-windows: copy
	if not exist $(OUTPUT_DIR) mkdir $(OUTPUT_DIR)
	go build -o $(OUTPUT_DIR)\digger-windows-amd64.exe -v .\cmd\digger
	go build -ldflags "$(LDARGS)" -o $(OUTPUT_DIR)\service_manager.exe -v .\cmd\service_manager

copy:
	if not exist $(OUTPUT_DIR) mkdir $(OUTPUT_DIR)
	if not exist $(OUTPUT_DIR)\templates mkdir $(OUTPUT_DIR)\templates
	copy /Y config.yaml $(OUTPUT_DIR)
	copy /Y sites.csv $(OUTPUT_DIR)
	copy /Y templates\email.html $(OUTPUT_DIR)\templates


build-all: build-windows

test:
	go test -v ./... -cover

clean:
	@echo Cleaning build artifacts...
	@if exist "$(OUTPUT_DIR)\*.exe" del /F /Q "$(OUTPUT_DIR)\*.exe"
	@if exist "$(OUTPUT_DIR)\*.exe~" del /F /Q "$(OUTPUT_DIR)\*.exe~"
	@if exist "$(OUTPUT_DIR)\digger-*" del /F /Q "$(OUTPUT_DIR)\digger-*"
	@if exist "$(OUTPUT_DIR)\config.yaml" del /F /Q "$(OUTPUT_DIR)\config.yaml"
	@if exist "$(OUTPUT_DIR)" rd /S /Q "$(OUTPUT_DIR)"
	@if exist "digger.exe" del /F /Q "digger.exe"
	@if exist "digger.log" del /F /Q "digger.log"
	@echo Clean completed.

run: build-windows
	.\$(OUTPUT_DIR)\digger-windows-amd64.exe

zip: clean build-windows
	@echo Zipping build artifacts...
	@if exist "$(OUTPUT_DIR)\digger-windows-amd64.zip" del /F /Q "$(OUTPUT_DIR)\digger-windows-amd64.zip"
	powershell -Command "Compress-Archive -Path $(OUTPUT_DIR)\* -DestinationPath $(OUTPUT_DIR)\digger-windows-amd64.zip"
	@echo Zip completed.