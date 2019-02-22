# MAINTAINER: David López <not4rent@gmail.com>

SHELL            = /bin/bash
MIGRATOR_DB_DSN ?= postgres://postgres@localhost/migrator?sslmode=disable

.PHONY: sanity-check
sanity-check: lint test

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	2>&1 MIGRATOR_DB_DSN="$(MIGRATOR_DB_DSN)" go test -tags="unit integration"