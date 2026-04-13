# update-go-mod-action

`update-go-mod-action` is a standalone custom GitHub Action that checks the latest stable Go release from `go.dev` and updates the `go` directive in `go.mod` when a repository is behind.

It is designed to stay small and composable:

- this action decides whether `go.mod` needs a Go version bump and edits the file
- a separate workflow step can create a pull request if a change was made

## Inputs

- `go-mod-path`: path to the `go.mod` file. Default: `go.mod`
- `update-toolchain`: when `true`, also updates an existing `toolchain go...` line to match. Default: `false`

## Outputs

- `changed`: `true` when the action modified the target file
- `previous-version`: the version originally declared by the `go` directive
- `current-version`: the version declared after the action runs
- `latest-version`: the latest stable Go version found from `go.dev`

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
        uses: faisal-memon/update-go-mod-action@main
        with:
          update-toolchain: "true"

      - name: Create pull request
        if: steps.update.outputs.changed == 'true'
        uses: peter-evans/create-pull-request@v7
        with:
          commit-message: Update Go version to ${{ steps.update.outputs.latest-version }}
          title: Update Go version to ${{ steps.update.outputs.latest-version }}
          body: |
            Updates the go.mod Go version to the latest stable release.
          branch: chore/update-go-${{ steps.update.outputs.latest-version }}
```

## Suggested Release Model

After publishing the repository, create a stable major tag such as `v1` so other repositories can use:

```yaml
uses: faisal-memon/update-go-mod-action@v1
```

You can move `v1` forward as you publish compatible updates.
