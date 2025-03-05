APP := digger
APP_VERSION := 0.1.0
APP_BUILD := $(shell git rev-parse --short HEAD)
APP_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X main.version=$(APP_VERSION) -X main.build=$(APP_BUILD) -X main.date=$(APP_DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o $(APP) main.go