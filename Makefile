SHELL := /bin/bash -e

.DEFAULT_GOAL := help

.PHONY: all
all: test lint ## Run all the tests and linters

.PHONY: clean
clean: ## Clean the working directory from binaries, coverage
	rm -f tmp/coverage/*

.PHONY: test
test: buildable test-unit test-examples benchmark test-flakiness ## Run all the tests (unit & benchmark)

.PHONY: buildable
buildable: ## Check if the project is buildable
	go build -v ./...

.PHONY: test-unit
test-unit: ## Run the unit tests
	mkdir -p tmp/coverage
	go test -v -timeout 10s -race -coverprofile=tmp/coverage/cover.out \
		./...

.PHONY: test-examples
test-examples: ## Run the testable examples
	mkdir -p tmp/coverage
	go test -v -run '^Example' -coverprofile=tmp/coverage/example.out \
		./...

.PHONY: benchmark
benchmark: ## Run the benchmarks
	mkdir -p tmp/coverage
	go test -v -run '^$$' -bench '^Benchmark' -coverprofile=tmp/coverage/benchmark.out \
		./...

.PHONY: test-flakiness
test-flakiness: ## Run the unit tests with a high count to ensure they are not flaky
	mkdir -p tmp/coverage
	# Yes, we really can run the tests 100000 times in just a few seconds
	go test -timeout 2m -count 100000 -failfast -skip '^Example' ./...

.PHONY: lint
lint: ## Run the linters
	pre-commit run --all-files

# Implements this pattern for autodocumenting Makefiles:
# https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
#
# Picks up all comments that start with a ## and are at the end of a target definition line.
help:
	@grep -h -E '^[0-9a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
