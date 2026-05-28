.PHONY: test clean

REMOTE     ?= origin
GO_VERSION ?= $(shell go env GOVERSION | sed 's/^go//')
LAST_TAG    = $(shell git ls-remote --tags --refs ${REMOTE} 'v*' | sed 's|.*refs/tags/||' | sort -V | tail -1)

all: clean tidy-check lint api-check cross-check test

.PHONY: lint
lint: ## Run linter
	go tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --timeout=15m ./...

.PHONY: test
test: ## Run tests
	go tool gotest.tools/gotestsum --junitfile=junit.xml -- -race -covermode=atomic -coverprofile=coverage.txt ./...

.PHONY: tidy-check
tidy-check: ## Check that go.mod and go.sum are tidy
	go mod tidy
	git diff --exit-code --name-status -- go.mod go.sum

.PHONY: cross-check
cross-check: ## Cross-compile for windows and darwin
	GOOS=windows go build ./...
	GOOS=darwin  go build ./...

.PHONY: api-check
api-check: BASE = $(if ${LAST_TAG},${LAST_TAG},none -version=v1.0.0)
api-check: ## Fail on breaking API changes vs the latest tag
	go tool gorelease -base=${BASE}

.PHONY: tag
tag: ## Tag commit using gorelease's suggested version
	@v="v1.0.0"; if [ -n "${LAST_TAG}" ]; then \
	  v=$$(go tool gorelease -base=${LAST_TAG} | tee /dev/stderr | awk '/^Suggested version:/ {print $$3; exit}'); \
	  test -n "$$v" || { echo "gorelease did not suggest a version" >&2; exit 1; }; \
	fi; \
	git tag "$$v" && echo "tagged $$v"

.PHONY: update-deps
update-deps: ## Update Go version, tools, and deps
	go mod edit -go=${GO_VERSION}
	go get $$(go mod edit -json | jq -r '[(.Tool[]?.Path), (.Require[]? | select(.Indirect | not) | .Path)] | map(. + "@latest") | .[]')
	go mod tidy

.PHONY: clean
clean: ## Clean files
	git clean -Xdf
