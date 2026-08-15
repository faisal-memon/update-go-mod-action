package updater

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateGoDirective(t *testing.T) {
	path := writeTempGoMod(t, "module example.com/app\n\ngo 1.25.0\n")

	result, err := update(Config{
		GoModPath: path,
		hooks:     versionHooks("1.26.0"),
	})
	require.NoError(t, err)

	assert.True(t, result.Changed)
	assert.Equal(t, "1.25.0", result.PreviousVersion)
	assert.Equal(t, "1.26.0", result.CurrentVersion)

	contents := readFile(t, path)
	assert.Contains(t, contents, "go 1.26.0\n")
}

func TestUpdateToolchainWhenRequested(t *testing.T) {
	path := writeTempGoMod(t, "module example.com/app\n\ngo 1.25.0\ntoolchain go1.25.0\n")

	_, err := update(Config{
		GoModPath:       path,
		UpdateToolchain: true,
		hooks:           versionHooks("1.26.0"),
	})
	require.NoError(t, err)

	contents := readFile(t, path)
	assert.Contains(t, contents, "toolchain go1.26.0\n")
}

func TestLeaveCurrentGoModUnchanged(t *testing.T) {
	original := "module example.com/app\n\ngo 1.26.0\n"
	path := writeTempGoMod(t, original)

	result, err := update(Config{
		GoModPath: path,
		hooks:     versionHooks("1.26.0"),
	})
	require.NoError(t, err)

	assert.False(t, result.Changed)
	assert.Equal(t, original, readFile(t, path))
}

func TestUpdatePreservesGoModPermissions(t *testing.T) {
	path := writeTempGoMod(t, "module example.com/app\n\ngo 1.25.0\n")
	require.NoError(t, os.Chmod(path, 0o600))

	_, err := update(Config{
		GoModPath: path,
		hooks:     versionHooks("1.26.0"),
	})
	require.NoError(t, err)

	fileInfo, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0o600), fileInfo.Mode().Perm())
}

func TestWriteGitHubOutputs(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "github_output")

	err := writeOutputs(outputPath, Result{
		Changed:         true,
		PreviousVersion: "1.25.0",
		CurrentVersion:  "1.26.0",
		LatestVersion:   "1.26.0",
	})
	require.NoError(t, err)

	contents := readFile(t, outputPath)
	want := "changed=true\nprevious-version=1.25.0\ncurrent-version=1.26.0\nlatest-version=1.26.0\n"
	assert.Equal(t, want, contents)
}

func TestVersionLessNormalizesGoVersions(t *testing.T) {
	less, err := versionLess("go1.25.9", "1.26.0")
	require.NoError(t, err)

	assert.True(t, less)
}

func TestVersionLessRejectsInvalidVersion(t *testing.T) {
	_, err := versionLess("not-a-version", "1.26.0")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Go version")
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
