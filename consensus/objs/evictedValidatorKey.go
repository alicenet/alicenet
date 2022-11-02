package objs

import (
	"bytes"

	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// EvictedValidatorKey stores DB metadata about an evicted validator
type EvictedValidatorKey struct {
	Prefix   []byte
	GroupKey []byte
	VAddress []byte
}

// MarshalBinary takes the AccusationKey object and returns
// the canonical byte slice
func (a *EvictedValidatorKey) MarshalBinary() ([]byte, error) {
	if a == nil {
		panic("EvictedValidatorKey.MarshalBinary; not initialized")
	}

	key := []byte{}
	key = append(key, utils.CopySlice(a.Prefix)...)
	key = append(key, []byte("|")...)
	key = append(key, utils.CopySlice(a.GroupKey)...)
	key = append(key, []byte("|")...)
	key = append(key, utils.CopySlice(a.VAddress)...)

	return key, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// EvictedValidatorKey object
func (a *EvictedValidatorKey) UnmarshalBinary(data []byte) error {
	if a == nil {
		panic("EvictedValidatorKey.UnmarshalBinary; not initialized")
	}
	splitData := bytes.Split(data, []byte("|"))
	if len(splitData) != 3 {
		return errorz.ErrCorrupt
	}
	//a.Prefix = splitData[0]
	copy(a.Prefix[:], splitData[0])
	copy(a.GroupKey[:], splitData[1])
	copy(a.VAddress[:], splitData[2])

	return nil
}

// MakeIterKey takes the EvictedValidatorKey object and returns
// the canonical byte slice without the groupKey and vAddress
func (a *EvictedValidatorKey) MakeIterKey() ([]byte, error) {
	if a == nil {
		panic("EvictedValidatorKey.MakeIterKey; not initialized")
	}

	key := []byte{}
	key = append(key, utils.CopySlice(a.Prefix)...)

	return key, nil
}

// MakeIterKey takes the EvictedValidatorKey object and returns
// the canonical byte slice without the vAddress
func (a *EvictedValidatorKey) MakeGroupIterKey() ([]byte, error) {
	if a == nil {
		panic("EvictedValidatorKey.MakeGroupIterKey; not initialized")
	}

	key := []byte{}
	key = append(key, utils.CopySlice(a.Prefix)...)
	key = append(key, []byte("|")...)
	key = append(key, utils.CopySlice(a.GroupKey)...)

	return key, nil
}
