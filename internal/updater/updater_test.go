package updater

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateGoDirective(t *testing.T) {
	path := writeTempGoMod(t, "module example.com/app\n\ngo 1.25.0\n")
	outputPath := filepath.Join(t.TempDir(), "github_output")

	err := Run(testConfig(path, outputPath, versionHooks("1.26.0")))
	require.NoError(t, err)

	contents := readFile(t, path)
	assert.Contains(t, contents, "go 1.26.0\n")
	assert.Equal(t, "changed=true\nprevious-version=1.25.0\nupdated-version=1.26.0\n", readFile(t, outputPath))
}

func TestUpdateToolchainWhenRequested(t *testing.T) {
	path := writeTempGoMod(t, "module example.com/app\n\ngo 1.25.0\ntoolchain go1.25.0\n")

	config := testConfig(path, "", versionHooks("1.26.0"))
	config.UpdateToolchain = true
	err := Run(config)
	require.NoError(t, err)

	contents := readFile(t, path)
	assert.Contains(t, contents, "toolchain go1.26.0\n")
}

func TestLeaveCurrentGoModUnchanged(t *testing.T) {
	original := "module example.com/app\n\ngo 1.26.0\n"
	path := writeTempGoMod(t, original)
	outputPath := filepath.Join(t.TempDir(), "github_output")

	err := Run(testConfig(path, outputPath, versionHooks("1.26.0")))
	require.NoError(t, err)

	assert.Equal(t, original, readFile(t, path))
	assert.Equal(t, "changed=false\n", readFile(t, outputPath))
}

func TestUpdatePreservesGoModPermissions(t *testing.T) {
	path := writeTempGoMod(t, "module example.com/app\n\ngo 1.25.0\n")
	require.NoError(t, os.Chmod(path, 0o600))

	err := Run(testConfig(path, "", versionHooks("1.26.0")))
	require.NoError(t, err)

	fileInfo, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
}

func TestWriteGitHubOutputs(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "github_output")

	err := writeOutputs(outputPath, true, "1.25.0", "1.26.0")
	require.NoError(t, err)

	contents := readFile(t, outputPath)
	want := "changed=true\nprevious-version=1.25.0\nupdated-version=1.26.0\n"
	assert.Equal(t, want, contents)
}

func TestWriteGitHubOutputsWithoutVersionsWhenUnchanged(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "github_output")

	err := writeOutputs(outputPath, false, "1.26.0", "1.26.0")
	require.NoError(t, err)

	contents := readFile(t, outputPath)
	assert.Equal(t, "changed=false\n", contents)
}

func writeTempGoMod(t *testing.T, contents string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "go.mod")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))

	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	contents, err := os.ReadFile(path)
	require.NoError(t, err)

	return string(contents)
}

func versionHooks(version string) hooks {
	return hooks{
		latestStableVersion: func() (string, error) {
			return version, nil
		},
	}
}

func testConfig(goModPath, outputPath string, hooks hooks) Config {
	return Config{
		GoModPath:    goModPath,
		GitHubOutput: outputPath,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
		hooks:        hooks,
	}
}
