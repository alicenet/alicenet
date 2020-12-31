package objs

import (
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	gUtils "github.com/MadBase/MadNet/utils"
)

// BlockHeaderHashIndexKey ...
type BlockHeaderHashIndexKey struct {
	Prefix    []byte
	BlockHash []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// BlockHeaderHashIndexKey object
func (b *BlockHeaderHashIndexKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if len(data) != (constants.HashLen + 2) {
		return errorz.ErrInvalid{}.New("Invalid BlockHeaderHashIndexKey")
	}
	b.Prefix = gUtils.CopySlice(data[0:2])
	b.BlockHash = gUtils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the BlockHeaderHashIndexKey object and returns
// the canonical byte slice
func (b *BlockHeaderHashIndexKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	if len(b.BlockHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("BlockHeaderHashIndexKey: invalid BlockHash")
	}
	if len(b.Prefix) != 2 {
		return nil, errorz.ErrInvalid{}.New("BlockHeaderHashIndexKey: invalid Prefix")
	}
	key := []byte{}
	Prefix := gUtils.CopySlice(b.Prefix)
	BlockHash := gUtils.CopySlice(b.BlockHash)
	key = append(key, Prefix...)
	key = append(key, BlockHash...)
	return key, nil
}
