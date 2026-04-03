APP_NAME := go-sling
VERSION := 1.0.0
BUILD_DIR := bin
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: all build run clean test build-all build-raspi build-linux build-mac build-windows

all: build

build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) .

run: build
	./$(BUILD_DIR)/$(APP_NAME)

build-raspi:
	GOOS=linux GOARCH=arm GOARM=7 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm7 .

build-linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 .

build-mac:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 .

build-windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe .

build-all: build-raspi build-linux build-mac build-windows

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

lint:
	golangci-lint run
