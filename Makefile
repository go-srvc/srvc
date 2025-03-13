.PHONY: test clean

all: clean lint test

.PHONY: lint
lint: ## Run linter
	go tool github.com/golangci/golangci-lint/cmd/golangci-lint run --timeout=15m ./...

.PHONY: test
test: ## Run tests
	go tool gotest.tools/gotestsum --junitfile=junit.xml -- -race -covermode=atomic -coverprofile=coverage.txt ./...

.PHONY: clean
clean: ## Clean files
	git clean -Xdf
