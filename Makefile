.PHONY: all build test clean install lint

VERSION ?= $(shell git describe --tags --always --dirty)
LDFLAGS := -X main.version=$(VERSION) -X main.buildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BINARY := sheetsync

all: test build

build:
	@echo "Building SheetSync $(VERSION)..."
	@mkdir -p dist
	go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY) ./cmd/sheetsync

build-all:
	@echo "Building SheetSync $(VERSION) for all platforms..."
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-amd64 ./cmd/sheetsync
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-darwin-arm64 ./cmd/sheetsync
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-windows-amd64.exe ./cmd/sheetsync
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/$(BINARY)-linux-amd64 ./cmd/sheetsync

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

clean:
	@echo "Cleaning..."
	@rm -rf dist/ coverage.txt

install: build
	@echo "Installing SheetSync..."
	@cp dist/$(BINARY) $(GOPATH)/bin/$(BINARY)

run: build
	./dist/$(BINARY)

.DEFAULT_GOAL := all