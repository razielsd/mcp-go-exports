# Project variables
BINARY_NAME=mcp-server

.PHONY: all build run test list clean help

all: build

## Build the binary
build:
	go build -o $(BINARY_NAME) .

## Run the server (expects JSON-RPC input via stdin)
run:
	go run .

## Run all tests
test:
	go test ./...

## List available tools in the MCP server
list:
	@echo '{"jsonrpc": "2.0", "method": "tools/list", "id": 1}' | go run .

## Launch MCP Inspector for interactive testing
inspector:
	npx @modelcontextprotocol/inspector go run .

## Clean build artifacts
clean:
	rm -f $(BINARY_NAME)

## Show this help message
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## ";} {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint:
	golangci-lint run  --config=.golangci.yml --timeout=360s ./...

.PHONY: fmt
fmt:
	gofumpt -w ./internal/
	gofumpt -w ./main.go
	gci write --skip-generated -s standard -s default -s "prefix(github.com/razielsd/mcp-go-exports)" -s blank -s dot --custom-order ./internal/..
	gci write --skip-generated -s standard -s default -s "prefix(github.com/razielsd/mcp-go-exports)" -s blank -s dot --custom-order ./main.go