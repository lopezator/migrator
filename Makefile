# MAINTAINER: David López <not4rent@gmail.com>

SHELL = /bin/bash

POSTGRES_URL = postgres://postgres:migrator@postgres:5432/migrator?sslmode=disable
MYSQL_URL    = root:migrator@tcp(mysql:3306)/migrator

.PHONY: setup-env
setup-env:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b . v1.44.2

.PHONY: prepare
prepare: setup-env mod-download

.PHONY: mod-download
mod-download:
	@echo "Running download..."
	go mod download

.PHONY: sanity-check
sanity-check: golangci-lint

.PHONY: golangci-lint
golangci-lint:
	@echo "Running lint..."
	./golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	2>&1 POSTGRES_URL="$(POSTGRES_URL)" MYSQL_URL="$(MYSQL_URL)" go test -v -tags="unit integration" -coverprofile=coverage.txt -covermode=atomic
