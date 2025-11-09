.PHONY: all build clean test proto install help

# Build variables
BINARY_NAME=meshtastic-mqtt-relay
CMD_DIR=cmd/meshtastic-mqtt-relay
PROTO_DIR=pkg/meshtastic

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
PROTOCMD=protoc

all: proto build

build: ## Build the binary
	$(GOBUILD) -v -o $(BINARY_NAME) ./$(CMD_DIR)

clean: ## Remove build artifacts
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

test: ## Run tests
	$(GOTEST) -v ./...

proto: ## Generate Go code from proto files
	$(PROTOCMD) --go_out=. --go_opt=paths=source_relative $(PROTO_DIR)/meshtastic.proto

install: build ## Install the binary
	install -m 755 $(BINARY_NAME) /usr/local/bin/

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

run: build ## Build and run the application
	./$(BINARY_NAME)

help: ## Display this help message
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-15s\033[0m %s\n", $$1, $$2}'
