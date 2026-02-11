.PHONY: test test-cover lint fmt vet ci integration-test integration-up integration-down clean help

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

## test: Run unit tests
test:
	go test -v -count=1 ./... -short

## test-cover: Run tests with coverage report
test-cover:
	go test -v -count=1 -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

## lint: Run golangci-lint (requires v2+)
lint:
	golangci-lint run ./...

## fmt: Check code formatting
fmt:
	@test -z "$$(gofmt -l .)" || (echo "Files need formatting:"; gofmt -l .; exit 1)

## vet: Run go vet
vet:
	go vet ./...

## ci: Run all CI checks (fmt, vet, lint, test)
ci: fmt vet lint test

## integration-up: Start Docker containers for integration tests
integration-up:
	cd integration && docker compose up -d --build
	cd integration && docker compose run --rm php-writer
	@echo "Memcached is ready with test data"

## integration-down: Stop Docker containers
integration-down:
	cd integration && docker compose down -v

## integration-test: Run integration tests (starts/stops Docker automatically)
integration-test: integration-up
	MEMCACHED_HOST=localhost MEMCACHED_PORT=11211 go test -v -count=1 -tags=integration ./integration/
	$(MAKE) integration-down

## clean: Remove build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean -testcache
