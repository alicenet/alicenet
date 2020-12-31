package objs

import (
	"github.com/MadBase/MadNet/errorz"
	gUtils "github.com/MadBase/MadNet/utils"
)

// StagedBlockHeaderKey ...
type StagedBlockHeaderKey struct {
	Prefix []byte
	Key    []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// StagedBlockHeaderKey object
func (b *StagedBlockHeaderKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if len(data) != 6 {
		return errorz.ErrInvalid{}.New("Invalid data for StagedBlockHeaderKey unmarshalling")
	}
	b.Prefix = gUtils.CopySlice(data[0:2])
	b.Key = gUtils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the StagedBlockHeaderKey object and returns the canonical
// byte slice
func (b *StagedBlockHeaderKey) MarshalBinary() ([]byte, error) {
	if b == nil || len(b.Prefix) != 2 || len(b.Key) != 4 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	Prefix := gUtils.CopySlice(b.Prefix)
	key = append(key, Prefix...)
	key = append(key, b.Key...)
	return key, nil
}
