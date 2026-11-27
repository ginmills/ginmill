SHELL := /usr/bin/env bash
.DEFAULT_GOAL := help

GO ?= go
GOLANGCI_LINT ?= $(shell command -v golangci-lint 2>/dev/null || echo "$$($(GO) env GOPATH)/bin/golangci-lint")

COVERAGE_DIR := .coverage
COVERAGE_FILE := $(COVERAGE_DIR)/coverage.out
COVERAGE_HTML := $(COVERAGE_DIR)/coverage.html
COVERAGE_MIN ?= 80.0

GO_PKGS := $(shell $(GO) list ./... 2>/dev/null)

.PHONY: help fmt fmt-check vet lint lint-fix test test-race test-coverage coverage-html coverage-check check clean

help: ## Show available targets
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make <target>\n\nTargets:\n"} /^[a-zA-Z0-9_-]+:.*?##/ {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

fmt: ## Format Go files via golangci-lint formatters (or gofmt fallback)
	@if [ -x "$(GOLANGCI_LINT)" ]; then \
		"$(GOLANGCI_LINT)" fmt ./...; \
	else \
		gofmt -s -w .; \
	fi

fmt-check: ## Verify Go files are properly formatted
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Unformatted Go files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet: ## Run go vet across all packages
	@if [ -z "$(GO_PKGS)" ]; then \
		echo "No Go packages found yet; skipping go vet."; \
	else \
		$(GO) vet ./...; \
	fi

lint: ## Run golangci-lint using .golangci.yml
	@if [ ! -x "$(GOLANGCI_LINT)" ]; then \
		echo "golangci-lint not found. Install v2 with:"; \
		echo "  go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
	@"$(GOLANGCI_LINT)" config verify
	@if [ -z "$(GO_PKGS)" ]; then \
		echo "No Go packages found yet; skipping golangci-lint run."; \
	else \
		"$(GOLANGCI_LINT)" run ./...; \
	fi

lint-fix: ## Run golangci-lint with auto-fix enabled
	@"$(GOLANGCI_LINT)" run --fix ./...

test: ## Run unit tests (fast feedback during TDD)
	@if [ -z "$(GO_PKGS)" ]; then \
		echo "No Go packages found yet; skipping go test."; \
	else \
		$(GO) test -count=1 ./...; \
	fi

test-race: ## Run full test suite with race detector (mandatory TDD GREEN gate)
	@if [ -z "$(GO_PKGS)" ]; then \
		echo "No Go packages found yet; skipping go test -race."; \
	else \
		$(GO) test -race -count=1 -v ./...; \
	fi

test-coverage: ## Run tests with race detector and generate coverage report
	@mkdir -p $(COVERAGE_DIR)
	@if [ -z "$(GO_PKGS)" ]; then \
		echo "No Go packages found yet; skipping coverage."; \
	else \
		$(GO) test -race -count=1 -covermode=atomic -coverprofile=$(COVERAGE_FILE) ./... && \
		$(GO) tool cover -func=$(COVERAGE_FILE); \
	fi

coverage-html: test-coverage ## Generate HTML coverage report in .coverage/coverage.html
	@if [ -f "$(COVERAGE_FILE)" ]; then \
		$(GO) tool cover -html=$(COVERAGE_FILE) -o $(COVERAGE_HTML) && \
		echo "HTML coverage report written to $(COVERAGE_HTML)"; \
	fi

coverage-check: test-coverage ## Enforce minimum total coverage (default 80.0%, override via COVERAGE_MIN)
	@if [ ! -f "$(COVERAGE_FILE)" ]; then \
		echo "No coverage file generated; skipping threshold check."; \
	else \
		total=$$($(GO) tool cover -func=$(COVERAGE_FILE) | awk '/^total:/ {gsub(/%/, "", $$3); print $$3}'); \
		echo "Total coverage: $${total}% (minimum: $(COVERAGE_MIN)%)"; \
		awk -v total="$$total" -v min="$(COVERAGE_MIN)" 'BEGIN { exit (total + 0 < min + 0) ? 1 : 0 }'; \
	fi

check: fmt-check vet lint test-coverage ## Full pre-archive / pre-PR verification gate
	@echo "All verification checks passed."

clean: ## Remove generated coverage and build artifacts
	@rm -rf $(COVERAGE_DIR) bin
