# Contributor Guide

## Repository Overview

- `action.yml` defines the public composite GitHub Action.
- `cmd/update-go-mod` contains the action entrypoint.
- `internal/updater` contains Go-version discovery, semantic-version handling, and `go.mod` updates.
- `.github/workflows` contains PR validation, release, and scheduled tooling-update workflows.
- `.versions` pins development-tool versions used by the Makefile.

## Development Commands

- `make test` runs all Go tests.
- `make lint` checks formatting, runs `go vet`, and runs the pinned golangci-lint binary.
- `make lint-fix` applies safe golangci-lint fixes.
- `make tidy` runs `go mod tidy`.

The first `make lint` or `make lint-fix` may download the pinned golangci-lint binary into `.tools`; subsequent runs reuse it.

## Change Guidelines

- Keep the action entrypoint and its inputs and outputs documented in `action.yml` and `README.md`.
- Use the standard library and existing dependencies where practical; semantic-version comparisons use the Go module already declared in `go.mod`.
- Pass the logger from the command entrypoint into updater code rather than using package-level logging.
- Keep network access for fetching the latest stable Go release isolated from file-update logic.
- Add or update unit tests with behavioral changes. Use `testify` assertions consistently with the existing tests.
- Run `gofmt` on changed Go files, then run `make lint` and `make test`.

## GitHub Actions

- Use `actions/setup-go` with the appropriate `go.mod` file when a workflow needs Go.
- The release workflow computes the next version with `pr-label-semver`, creates the exact GitHub release with generated notes, then moves the floating major tag.
- The scheduled tooling workflow uses this action to update this repository's Go version and the separate golangci-lint action to update `.versions`.
- Dependabot is used to keep GitHub Actions and Go dependencies current.

## Pull Requests

- Start from the latest `main` before creating a feature branch.
- Keep PRs focused and describe workflow or release-behavior changes explicitly.
- Before requesting review, run `make lint` and `make test` and include any limitations in the PR description.
