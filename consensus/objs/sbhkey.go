package objs

import (
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
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
		return errorz.ErrInvalid{}.New("StagedBlockHeaderKey.UnmarshalBinary; sbhk not initialized")
	}
	if len(data) != 6 {
		return errorz.ErrInvalid{}.New("StagedBlockHeaderKey.UnmarshalBinary; incorrect state length")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	b.Key = utils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the StagedBlockHeaderKey object and returns the canonical
// byte slice
func (b *StagedBlockHeaderKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("StagedBlockHeaderKey.MarshalBinary; sbhk not initialized")
	}
	if len(b.Prefix) != 2 {
		return nil, errorz.ErrInvalid{}.New("StagedBlockHeaderKey.MarshalBinary; incorrect prefix length")
	}
	if len(b.Key) != 4 {
		return nil, errorz.ErrInvalid{}.New("StagedBlockHeaderKey.MarshalBinary; incorrect key length")
	}
	key := []byte{}
	Prefix := utils.CopySlice(b.Prefix)
	key = append(key, Prefix...)
	key = append(key, b.Key...)
	return key, nil
}
