.PHONY: test clean

GO_VERSION ?= $(shell go env GOVERSION | sed 's/^go//')

all: clean lint test

.PHONY: lint
lint: ## Run linter
	go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --timeout=15m ./...

.PHONY: test
test: ## Run tests
	go tool gotest.tools/gotestsum --junitfile=junit.xml -- -race -covermode=atomic -coverprofile=coverage.txt ./...

.PHONY: update-deps
update-deps: ## Update Go version, tools, and deps
	go mod edit -go=${GO_VERSION}
	go get -u -t ./...
	go get -u tool
	go mod tidy

.PHONY: clean
clean: ## Clean files
	git clean -Xdf
