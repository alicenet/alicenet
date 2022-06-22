package objs

import (
	"bytes"
	"encoding/hex"

	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

//# index historic by by groupKey|height|round|vAddr
//# index current by by groupKey|vkey

// RoundStateCurrentKey ...
type RoundStateCurrentKey struct {
	Prefix   []byte
	GroupKey []byte
	VAddr    []byte
}

// MarshalBinary takes the RoundStateCurrentKey object and returns
// the canonical byte slice
func (b *RoundStateCurrentKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("RoundStateCurrentKey.MarshalBinary; rsck not initialized")
	}
	key := []byte{}
	Prefix := utils.CopySlice(b.Prefix)
	GroupKey := make([]byte, hex.EncodedLen(len(b.GroupKey)))
	VAddr := make([]byte, hex.EncodedLen(len(b.VAddr)))
	_ = hex.Encode(GroupKey, b.GroupKey)
	_ = hex.Encode(VAddr, b.VAddr)
	key = append(key, Prefix...)
	key = append(key, []byte("|")...)
	key = append(key, GroupKey...)
	key = append(key, []byte("|")...)
	key = append(key, VAddr...)
	return key, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// RoundStateCurrentKey object
func (b *RoundStateCurrentKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("RoundStateCurrentKey.UnmarshalBinary; rsck not initialized")
	}
	splitData := bytes.Split(data, []byte("|"))
	if len(splitData) != 3 {
		return errorz.ErrCorrupt
	}
	b.Prefix = splitData[0]
	GroupKey := make([]byte, hex.DecodedLen(len(splitData[1])))
	_, err := hex.Decode(GroupKey, splitData[1])
	if err != nil {
		return err
	}
	b.GroupKey = GroupKey
	VAddr := make([]byte, hex.DecodedLen(len(splitData[2])))
	_, err = hex.Decode(VAddr, splitData[2])
	if err != nil {
		return err
	}
	b.VAddr = VAddr
	return nil
}
