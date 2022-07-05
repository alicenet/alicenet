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
	UUID   []byte
}

// MarshalBinary takes the AccusationKey object and returns
// the canonical byte slice
func (a *AccusationKey) MarshalBinary() ([]byte, error) {
	if a == nil {
		return nil, errorz.ErrInvalid{}.New("AccusationKey.MarshalBinary; accusation not initialized")
	}

	key := []byte{}
	Prefix := utils.CopySlice(a.Prefix)
	uuid := make([]byte, hex.EncodedLen(len(a.UUID)))
	_ = hex.Encode(uuid, a.UUID)
	key = append(key, Prefix...)
	key = append(key, []byte("|")...)
	key = append(key, uuid...)

	return key, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// AccusationKey object
func (a *AccusationKey) UnmarshalBinary(data []byte) error {
	if a == nil {
		return errorz.ErrInvalid{}.New("AccusationKey.UnmarshalBinary; accusation not initialized")
	}
	splitData := bytes.Split(data, []byte("|"))
	if len(splitData) != 4 {
		return errorz.ErrCorrupt
	}
	a.Prefix = splitData[0]
	UUID := make([]byte, hex.DecodedLen(len(splitData[1])))
	_, err := hex.Decode(UUID, splitData[1])
	if err != nil {
		return err
	}
	a.UUID = UUID

	return nil
}

// MakeIterKey takes the AccusationKey object and returns
// the canonical byte slice without the UUID
func (a *AccusationKey) MakeIterKey() ([]byte, error) {
	if a == nil {
		return nil, errorz.ErrInvalid{}.New("AccusationKey.MakeIterKey; accusation not initialized")
	}

	key := []byte{}
	Prefix := utils.CopySlice(a.Prefix)
	key = append(key, Prefix...)

	return key, nil
}
