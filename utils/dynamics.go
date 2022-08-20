package utils

import "github.com/alicenet/alicenet/bridge/bindings"

func CompareCanonicalVersion(newVersion bindings.CanonicalVersion) (bool, bool, bool, bindings.CanonicalVersion) {
	localVersion := GetLocalVersion()
	return newVersion.Major > localVersion.Major,
		newVersion.Major == localVersion.Major && newVersion.Minor > localVersion.Minor,
		newVersion.Major == localVersion.Major && newVersion.Minor == localVersion.Minor && newVersion.Patch > localVersion.Patch,
		localVersion
}

func GetLocalVersion() bindings.CanonicalVersion {
	return bindings.CanonicalVersion{
		Major:      0,
		Minor:      0,
		Patch:      0,
		BinaryHash: [32]byte{},
	}
}
