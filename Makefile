.PHONY: build clean run tidy
APP := digger
APP_VERSION := 0.1.0
APP_BUILD := $(shell git rev-parse --short HEAD)
APP_DATE := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -X main.version=$(APP_VERSION) -X main.build=$(APP_BUILD) -X main.date=$(APP_DATE)

build: clean tidy
	go build -ldflags "$(LDFLAGS)" -o $(APP) *.go

run: build
	./$(APP)

clean:
	rm -f $(APP)

tidy:
	go mod tidy