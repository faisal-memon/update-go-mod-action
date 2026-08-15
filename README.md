# update-go-mod-action

Keeps `go.mod` up to date with the latest stable Go release from `go.dev`.

The action sets up Go with `actions/setup-go` and runs its Go implementation directly, so callers do not need to add a separate Go setup step before using it.

## Example

```yaml
name: Update Go Version

on:
  workflow_dispatch:
  schedule:
    - cron: "0 9 * * 1"

jobs:
  update-go:
    runs-on: ubuntu-latest
    permissions:
      contents: write
      pull-requests: write

    steps:
      - name: Check out repository
        uses: actions/checkout@v4

      - name: Update go.mod
        id: update
        uses: faisal-memon/update-go-mod-action@v1
        with:
          update-toolchain: "true"

      - name: Create pull request
        if: steps.update.outputs.changed == 'true'
        uses: peter-evans/create-pull-request@v7
        with:
          commit-message: Update Go version to ${{ steps.update.outputs.updated-version }}
          title: Update Go version to ${{ steps.update.outputs.updated-version }}
          body: |
            Updates the go.mod Go version to the latest stable release.
          branch: chore/update-go-${{ steps.update.outputs.updated-version }}
```

## Inputs

- `go-mod-path`: path to the `go.mod` file. Default: `go.mod`
- `update-toolchain`: when `true`, also updates an existing `toolchain go...` line to match. Default: `false`

## Outputs

- `changed`: `true` when the action modified the target file
- `previous-version`: the version originally declared by the `go` directive
- `updated-version`: the version declared by the `go` directive after the action runs
