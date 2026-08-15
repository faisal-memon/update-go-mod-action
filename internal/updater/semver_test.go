package updater

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSemverCompareNormalizesGoVersions(t *testing.T) {
	comparison, err := semverCompare("go1.25.9", "1.26.0")
	require.NoError(t, err)

	assert.Negative(t, comparison)
}

func TestSemverCompareRejectsInvalidVersion(t *testing.T) {
	_, err := semverCompare("not-a-version", "1.26.0")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid Go version")
}
