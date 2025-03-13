.PHONY: test clean

all: test

test: clean
	go tool gotest.tools/gotestsum -- -coverprofile=coverage.txt ./...

clean:
	git clean -Xdf
