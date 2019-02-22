# MAINTAINER: David López <not4rent@gmail.com>

SHELL = /bin/bash

.PHONY: setup-env
setup-dev:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.12.5

.PHONY: sanity-check
sanity-check: download lint test

.PHONY: download
download:
	@echo "Running download..."
	go mod download

.PHONY: lint
lint:
	@echo "Running lint..."
	golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	2>&1 go test -tags="unit integration"