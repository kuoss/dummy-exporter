VERSION := dev
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

.PHONY: docker
docker: docker-build docker-run

.PHONY: docker-build
docker-build:
	docker build -t dummy-exporter:$(VERSION) --build-arg VERSION=$(VERSION) .

.PHONY: docker-run
docker-run:
	docker rm -f dummy-exporter
	docker run -d -p 9100:9100 --name dummy-exporter dummy-exporter:$(VERSION)
	sleep 1
	curl http://localhost:9100/metrics

.PHONY: docker-rm
docker-rm:
	docker rm -f dummy-exporter || true
