.PHONY: build mk-bin-dir build-api build-cli build-mcp deps clean fmt vet test check setup init

build: mk-bin-dir build-api build-cli build-mcp

mk-bin-dir:
	@mkdir -p bin

build-api:
	go build -o bin/kbnavt-api cmd/api/main.go

build-cli:
	go build -o bin/kbnavt-cli cmd/cli/main.go

build-mcp:
	go build -o bin/kbnavt-mcp cmd/mcp/main.go

deps:
	go mod download
	go mod tidy

clean:
	go clean
	@rm -rf bin

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

init: deps setup
	@echo "Project inited!"
