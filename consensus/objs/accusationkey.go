package objs

import (
	"bytes"
	"encoding/hex"

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
	id := make([]byte, hex.EncodedLen(len(a.ID)))
	_ = hex.Encode(id, a.ID[:])
	key = append(key, Prefix...)
	key = append(key, []byte("|")...)
	key = append(key, id...)

	return key, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// AccusationKey object
func (a *AccusationKey) UnmarshalBinary(data []byte) error {
	if a == nil {
		panic("AccusationKey.UnmarshalBinary; accusation not initialized")
	}
	splitData := bytes.Split(data, []byte("|"))
	if len(splitData) != 4 {
		return errorz.ErrCorrupt
	}
	a.Prefix = splitData[0]
	id := make([]byte, hex.DecodedLen(len(splitData[1])))
	_, err := hex.Decode(id, splitData[1])
	if err != nil {
		return err
	}
	copy(a.ID[:], id)

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
