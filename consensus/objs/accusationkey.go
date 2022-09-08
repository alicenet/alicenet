package objs

import (
	"bytes"

	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// AccusationKey stores DB metadata about an accusation
type AccusationKey struct {
	Prefix []byte
	ID     [32]byte
}

// MarshalBinary takes the AccusationKey object and returns
// the canonical byte slice
func (a *AccusationKey) MarshalBinary() ([]byte, error) {
	if a == nil {
		panic("AccusationKey.MarshalBinary; accusation not initialized")
	}

	key := []byte{}
	Prefix := utils.CopySlice(a.Prefix)
	key = append(key, Prefix...)
	key = append(key, []byte("|")...)
	key = append(key, a.ID[:]...)

	return key, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// AccusationKey object
func (a *AccusationKey) UnmarshalBinary(data []byte) error {
	if a == nil {
		panic("AccusationKey.UnmarshalBinary; accusation not initialized")
	}
	splitData := bytes.Split(data, []byte("|"))
	if len(splitData) != 2 {
		return errorz.ErrCorrupt
	}
	a.Prefix = splitData[0]
	copy(a.ID[:], splitData[1])

	return nil
}

// MakeIterKey takes the AccusationKey object and returns
// the canonical byte slice without the UUID
func (a *AccusationKey) MakeIterKey() ([]byte, error) {
	if a == nil {
		panic("AccusationKey.MakeIterKey; accusation not initialized")
	}

	key := []byte{}
	Prefix := utils.CopySlice(a.Prefix)
	key = append(key, Prefix...)

	return key, nil
}
