# update-go-mod-action

Simple action to keep `go.mod` up to date with the latest stable Go release from `go.dev`.

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
        uses: faisal-memon/update-go-mod-action@v2
        with:
          update-toolchain: "true"

      - name: Create pull request
        if: steps.update.outputs.changed == 'true'
        uses: peter-evans/create-pull-request@v7
        with:
          commit-message: Update Go version to ${{ steps.update.outputs.updated-version }}
          title: Update Go version to ${{ steps.update.outputs.updated-version }}
          body: go.mod Go version: `${{ steps.update-go.outputs.previous-version }}` -> `${{ steps.update-go.outputs.updated-version }}`.
          branch: chore/update-go-${{ steps.update.outputs.updated-version }}
```

## Inputs

| Argument | Default | Explanation |
| --- | --- | --- |
|`go-mod-path`| `go.mod`| path to the `go.mod` file |
|`update-toolchain`| `false` | When `true`, also updates an existing `toolchain go...` line to match |

## Outputs

| Value | Explanation |
| --- | --- |
|`changed` | `true` when the action modified the target file |
|`previous-version` | the version originally declared by the `go` directive when `changed` is `true` |
|`updated-version` | the version written to the `go` directive when `changed` is `true`|
