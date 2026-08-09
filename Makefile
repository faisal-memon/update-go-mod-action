GO ?= go
GOFMT ?= gofmt

.PHONY: lint test

lint:
	@test -z "$$($(GOFMT) -l .)"
	@$(GO) vet ./...

test:
	@$(GO) test ./...
