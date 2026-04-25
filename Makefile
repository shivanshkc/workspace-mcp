SHELL=/usr/bin/env bash

application_name		= workspace-mcp
application_binary_path	= /Users/skuchcha/NewPersonal/workspace-mcp/bin/workspace-mcp
default_config_path		= /etc/workspace-mcp/config.shivanshbox.json

# Builds the MCP server and adds it to Claude Code.
build:
	@echo "+$@"
	@go build -o bin/$(application_name) cmd/$(application_name)/main.go

claude: tidy build
	@claude mcp remove $(application_name) || true
	@claude mcp add $(application_name) -- $(application_binary_path) -config $(default_config_path)
	@claude mcp list

# Tests the whole project.
test:
	@echo "+$@"
	@CGO_ENABLED=1 go test -race -coverprofile=coverage.out -covermode=atomic ./...

# Runs the "go mod tidy" command.
tidy:
	@echo "+$@"
	@go mod tidy

# Runs golang-ci-lint over the project.
lint:
	@echo "+$@"
	@golangci-lint run
