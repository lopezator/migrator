.PHONY: sanity-check
sanity-check: golangci-lint

.PHONY: golangci-lint
golangci-lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...