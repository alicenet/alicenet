package utils

import (
	"errors"
	"strconv"
	"strings"

	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/config"
)

var (
	ErrDevelopmentVersion  = errors.New("not tagged build, running in developer mode")
	ErrUnexpectedTagFormat = errors.New("unexpected version format. expected vX.Y.Z[-optional]")
)

func CompareCanonicalVersion(newVersion bindings.CanonicalVersion) (bool, bool, bool, bindings.CanonicalVersion, error) {
	localVersion, err := GetLocalVersion()
	return newVersion.Major > localVersion.Major,
		newVersion.Major == localVersion.Major && newVersion.Minor > localVersion.Minor,
		newVersion.Major == localVersion.Major && newVersion.Minor == localVersion.Minor && newVersion.Patch > localVersion.Patch,
		localVersion, err
}

func GetLocalVersion() (bindings.CanonicalVersion, error) {
	var canonicalVersion bindings.CanonicalVersion

	version := config.Configuration.Version
	if version == "dev" {
		return canonicalVersion, ErrDevelopmentVersion
	}

	version = strings.TrimPrefix(version, "v")
	version, _, _ = strings.Cut(version, "-")
	versionParts := strings.Split(version, ".")

	if len(versionParts) != 3 {
		return canonicalVersion, ErrUnexpectedTagFormat
	}

	major, err := strconv.ParseUint(versionParts[0], 10, 32)
	if err != nil {
		return canonicalVersion, err
	}

	minor, err := strconv.ParseUint(versionParts[1], 10, 32)
	if err != nil {
		return canonicalVersion, err
	}

	patch, err := strconv.ParseUint(versionParts[2], 10, 32)
	if err != nil {
		return canonicalVersion, err
	}

	canonicalVersion.Major = uint32(major)
	canonicalVersion.Minor = uint32(minor)
	canonicalVersion.Patch = uint32(patch)

	return canonicalVersion, nil
}
