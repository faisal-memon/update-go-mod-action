package updater

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// semverCompare compares two Go version strings using semantic version rules.
func semverCompare(left, right string) (int, error) {
	leftVersion, err := toSemver(left)
	if err != nil {
		return 0, err
	}

	rightVersion, err := toSemver(right)
	if err != nil {
		return 0, err
	}

	return semver.Compare(leftVersion, rightVersion), nil
}

// toSemver normalizes a Go version string into the format required by x/mod/semver.
func toSemver(version string) (string, error) {
	cleaned := strings.TrimSpace(version)
	if cleaned == "" {
		return "", fmt.Errorf("empty Go version")
	}

	cleaned = strings.TrimPrefix(cleaned, "go")
	if !strings.HasPrefix(cleaned, "v") {
		cleaned = "v" + cleaned
	}

	if !semver.IsValid(cleaned) {
		return "", fmt.Errorf("invalid Go version %q", version)
	}

	return cleaned, nil
}
