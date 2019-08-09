# MAINTAINER: David López <not4rent@gmail.com>

SHELL = /bin/bash

GOPROXY      = https://proxy.golang.org
POSTGRES_URL = postgres://postgres@postgres:5432/migrator?sslmode=disable
MYSQL_URL    = root:mysql@tcp(mysql:3306)/migrator

.PHONY: setup-env
setup-env:
	GO111MODULE=off go get -u github.com/golangci/golangci-lint/cmd/golangci-lint
	GO111MODULE=off go get -u github.com/mjibson/esc

.PHONY: esc-gen
esc-gen:
	esc -pkg migrator -private -include=".*.sql$$" -modtime="0" testdata > sql.go

.PHONY: prepare
prepare: setup-env mod-download

.PHONY: mod-download
mod-download:
	@echo "Running download..."
	GOPROXY="$(GOPROXY)" go mod download

.PHONY: sanity-check
sanity-check: golangci-lint

.PHONY: golangci-lint
golangci-lint:
	@echo "Running lint..."
	golangci-lint run ./...

.PHONY: test
test:
	@echo "Running tests..."
	2>&1 POSTGRES_URL="$(POSTGRES_URL)" MYSQL_URL="$(MYSQL_URL)" go test -v -tags="unit integration" -coverprofile=coverage.txt -covermode=atomic
