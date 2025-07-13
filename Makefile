GOLANGCI_LINT_VERSION := v2.1.6

.PHONY: checks
checks: lint test

.PHONY: lint
lint: ./bin/golangci-lint
	./bin/golangci-lint run

./bin/golangci-lint:
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
		| sh -s -- -b ./bin/ $(GOLANGCI_LINT_VERSION)
	./bin/golangci-lint --version

.PHONY: test
test:
	go test -v ./...

.PHONY: cover
cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
