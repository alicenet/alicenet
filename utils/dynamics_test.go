package utils

import (
	"testing"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompareOnlyMajorUpdateCanonicalVersion(t *testing.T) {
	t.Parallel()
	config.Configuration.Version = "v0.0.0"
	updatedVersion, err := GetLocalVersion()
	require.Nil(t, err)

	updatedVersion.Major += 1
	isMajorUpdate, isMinorUpdate, isPatchUpdate, localVersion, err := CompareCanonicalVersion(updatedVersion)
	require.Nil(t, err)

	assert.Equal(t, isMajorUpdate, true)
	assert.Equal(t, isMinorUpdate, false)
	assert.Equal(t, isPatchUpdate, false)

	version, err := GetLocalVersion()
	require.Nil(t, err)
	assert.Equal(t, localVersion, version)
}

func TestCompareOnlyMinorUpdateCanonicalVersion(t *testing.T) {
	t.Parallel()
	config.Configuration.Version = "v0.0.0"
	updatedVersion, err := GetLocalVersion()
	require.Nil(t, err)

	updatedVersion.Minor += 1
	IsMajorUpdate, IsMinorUpdate, IsPatchUpdate, localVersion, err := CompareCanonicalVersion(updatedVersion)
	require.Nil(t, err)

	assert.Equal(t, IsMajorUpdate, false)
	assert.Equal(t, IsMinorUpdate, true)
	assert.Equal(t, IsPatchUpdate, false)

	version, err := GetLocalVersion()
	require.Nil(t, err)
	assert.Equal(t, localVersion, version)
}

func TestCompareOnlyPatchUpdateCanonicalVersion(t *testing.T) {
	t.Parallel()
	config.Configuration.Version = "v0.0.0"
	updatedVersion, err := GetLocalVersion()
	require.Nil(t, err)

	updatedVersion.Patch += 1
	IsMajorUpdate, IsMinorUpdate, IsPatchUpdate, localVersion, err := CompareCanonicalVersion(updatedVersion)
	require.Nil(t, err)
	assert.Equal(t, IsMajorUpdate, false)
	assert.Equal(t, IsMinorUpdate, false)
	assert.Equal(t, IsPatchUpdate, true)

	version, err := GetLocalVersion()
	require.Nil(t, err)
	assert.Equal(t, localVersion, version)
}

func TestCompareMixUpdateCanonicalVersion(t *testing.T) {
	t.Parallel()
	config.Configuration.Version = "v0.0.0"
	updatedVersion, err := GetLocalVersion()
	require.Nil(t, err)

	updatedVersion.Minor += 1
	updatedVersion.Patch += 100
	IsMajorUpdate, IsMinorUpdate, IsPatchUpdate, localVersion, err := CompareCanonicalVersion(updatedVersion)
	require.Nil(t, err)
	assert.Equal(t, IsMajorUpdate, false)
	assert.Equal(t, IsMinorUpdate, true)
	assert.Equal(t, IsPatchUpdate, false)

	version, err := GetLocalVersion()
	require.Nil(t, err)
	assert.Equal(t, localVersion, version)
}

func TestCompareMixUpdateCanonicalVersion2(t *testing.T) {
	t.Parallel()
	config.Configuration.Version = "v0.0.0"
	updatedVersion, err := GetLocalVersion()
	require.Nil(t, err)

	updatedVersion.Major += 1
	updatedVersion.Minor += 100
	updatedVersion.Patch += 100
	IsMajorUpdate, IsMinorUpdate, IsPatchUpdate, localVersion, err := CompareCanonicalVersion(updatedVersion)
	require.Nil(t, err)
	assert.Equal(t, IsMajorUpdate, true)
	assert.Equal(t, IsMinorUpdate, false)
	assert.Equal(t, IsPatchUpdate, false)

	version, err := GetLocalVersion()
	require.Nil(t, err)
	assert.Equal(t, localVersion, version)
}

func TestGetLocalVersion(t *testing.T) {
	// todo: fix this once we have the logic to get the canonical version from the binary.
	config.Configuration.Version = "v1.2.3"
	version, err := GetLocalVersion()
	require.Nil(t, err)
	assert.Equal(t, version, bindings.CanonicalVersion{
		Major:      1,
		Minor:      2,
		Patch:      3,
		BinaryHash: [32]byte{},
	})
}
