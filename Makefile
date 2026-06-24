BINARY := drishti
SHELL := /usr/bin/env bash
GOBIN := $(shell go env GOPATH)/bin

.PHONY: build build-ui fmt lint test cover run
build:
	go build -ldflags "-X main.version=$(shell git describe --tags --always 2>/dev/null || echo dev)" -o $(BINARY) ./cmd/drishti
build-ui:
	cd web && npm install && npm run build
fmt:
	gofmt -w .
	command -v goimports >/dev/null && goimports -w . || true
lint:
	golangci-lint run ./...
# test runs the suite AND feeds results to TDD Guard via the tdd-guard-go
# reporter. pipefail keeps go test's real exit code (the reporter always exits 0).
test:
	set -o pipefail; go test -json ./... -race 2>&1 | $(GOBIN)/tdd-guard-go -project-root $(CURDIR)
cover:
	go test ./... -coverprofile=cover.out && go tool cover -func=cover.out
run: build
	./$(BINARY)
