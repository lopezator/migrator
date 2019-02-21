# MAINTAINER: David López <not4rent@gmail.com>

.PHONY: sanity-check
sanity-check: lint test

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	go test -tags="$(TEST_LEVELS)"