# MAINTAINER: David López <not4rent@gmail.com>

SHELL = /bin/bash

GOPROXY      = https://athens.azurefd.net
POSTGRES_URL = postgres://postgres@localhost:2345/migrator?sslmode=disable
MARIA_DB_URL = root:mariadb@tcp(localhost:6033)/migrator

.PHONY: setup-env
setup-dev:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH)/bin v1.12.5

.PHONY: sanity-check
sanity-check: mod-download golangci-lint

.PHONY: mod-download
mod-download:
	@echo "Running download..."
	go mod download GOPROXY="$(GOPROXY)"

.PHONY: golangci-lint
golangci-lint:
	@echo "Running lint..."
	golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	2>&1 POSTGRES_URL="$(POSTGRES_URL)" MARIA_DB_URL="$(MARIA_DB_URL)" go test -tags="unit integration"