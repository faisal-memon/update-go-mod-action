GO ?= go
GOFMT ?= gofmt
CURL ?= curl

include .versions

TOOLS_DIR := .tools
GOLANGCI_LINT_DIR := $(TOOLS_DIR)/golangci-lint/$(GOLANGCI_LINT_VERSION)
GOLANGCI_LINT_BIN := $(GOLANGCI_LINT_DIR)/golangci-lint

.PHONY: lint lint-fix test

lint: $(GOLANGCI_LINT_BIN)
	@test -z "$$($(GOFMT) -l .)"
	@$(GO) vet ./...
	@$(GOLANGCI_LINT_BIN) run ./...

lint-fix: $(GOLANGCI_LINT_BIN)
	@$(GOLANGCI_LINT_BIN) run --fix ./...

test:
	@$(GO) test ./...

$(GOLANGCI_LINT_BIN):
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@rm -rf "$(dir $(GOLANGCI_LINT_DIR))"
	@mkdir -p "$(GOLANGCI_LINT_DIR)"
	@$(CURL) -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$(GOLANGCI_LINT_DIR)" "$(GOLANGCI_LINT_VERSION)"
