package utils

import (
	"testing"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/stretchr/testify/assert"
)

func TestCompareOnlyMajorUpdateCanonicalVersion(t *testing.T) {
	t.Parallel()
	updatedVersion := GetLocalVersion()
	updatedVersion.Major += 1
	isMajorUpdate, isMinorUpdate, isPatchUpdate, localVersion := CompareCanonicalVersion(updatedVersion)
	assert.Equal(t, isMajorUpdate, true)
	assert.Equal(t, isMinorUpdate, false)
	assert.Equal(t, isPatchUpdate, false)
	assert.Equal(t, localVersion, GetLocalVersion())
}

func TestCompareOnlyMinorUpdateCanonicalVersion(t *testing.T) {
	t.Parallel()
	updatedVersion := GetLocalVersion()
	updatedVersion.Minor += 1
	IsMajorUpdate, IsMinorUpdate, IsPatchUpdate, localVersion := CompareCanonicalVersion(updatedVersion)
	assert.Equal(t, IsMajorUpdate, false)
	assert.Equal(t, IsMinorUpdate, true)
	assert.Equal(t, IsPatchUpdate, false)
	assert.Equal(t, localVersion, GetLocalVersion())
}

func TestCompareOnlyPatchUpdateCanonicalVersion(t *testing.T) {
	t.Parallel()
	updatedVersion := GetLocalVersion()
	updatedVersion.Patch += 1
	IsMajorUpdate, IsMinorUpdate, IsPatchUpdate, localVersion := CompareCanonicalVersion(updatedVersion)
	assert.Equal(t, IsMajorUpdate, false)
	assert.Equal(t, IsMinorUpdate, false)
	assert.Equal(t, IsPatchUpdate, true)
	assert.Equal(t, localVersion, GetLocalVersion())
}

func TestCompareMixUpdateCanonicalVersion(t *testing.T) {
	t.Parallel()
	updatedVersion := GetLocalVersion()
	updatedVersion.Minor += 1
	updatedVersion.Patch += 100
	IsMajorUpdate, IsMinorUpdate, IsPatchUpdate, localVersion := CompareCanonicalVersion(updatedVersion)
	assert.Equal(t, IsMajorUpdate, false)
	assert.Equal(t, IsMinorUpdate, true)
	assert.Equal(t, IsPatchUpdate, false)
	assert.Equal(t, localVersion, GetLocalVersion())
}

func TestCompareMixUpdateCanonicalVersion2(t *testing.T) {
	t.Parallel()
	updatedVersion := GetLocalVersion()
	updatedVersion.Major += 1
	updatedVersion.Minor += 100
	updatedVersion.Patch += 100
	IsMajorUpdate, IsMinorUpdate, IsPatchUpdate, localVersion := CompareCanonicalVersion(updatedVersion)
	assert.Equal(t, IsMajorUpdate, true)
	assert.Equal(t, IsMinorUpdate, false)
	assert.Equal(t, IsPatchUpdate, false)
	assert.Equal(t, localVersion, GetLocalVersion())
}

func TestGetLocalVersion(t *testing.T) {
	// todo: fix this once we have the logic to get the canonical version from the binary
	assert.Equal(t, GetLocalVersion(), bindings.CanonicalVersion{
		Major:      0,
		Minor:      0,
		Patch:      0,
		BinaryHash: [32]byte{},
	})
}
