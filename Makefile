.PHONY: build build-api build-mcp run test clean deps swagger


build: build-api build-cli build-mcp

build-api:
	go build -o kbnavt-api cmd/api/main.go

build-cli:
	go build -o kbnavt-cli cmd/cli/main.go

build-mcp:
	go build -o kbnavt-mcp cmd/mcp/main.go

deps:
	go mod download
	go mod tidy

swagger:
	swag init -d cmd/api/

clean:
	go clean
	rm -f kbnavt-api rbnavt-mcp

install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest

fmt:
	go fmt ./...

vet:
	go vet ./...

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html

check: fmt vet test

setup:
	mkdir -p data

init: deps install-tools swagger setup
	@echo "Project inited!"
