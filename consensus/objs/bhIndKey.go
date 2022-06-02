package objs

import (
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
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
		return errorz.ErrInvalid{}.New("BlockHeaderHashIndexKey.UnmarshalBinary; bhhik not initialized")
	}
	if len(data) != (constants.HashLen + 2) {
		return errorz.ErrInvalid{}.New("BlockHeaderHashIndexKey.UnmarshalBinary; incorrect state length")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	b.BlockHash = utils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the BlockHeaderHashIndexKey object and returns
// the canonical byte slice
func (b *BlockHeaderHashIndexKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("BlockHeaderHashIndexKey.MarshalBinary; bhhik not initialized")
	}
	if len(b.BlockHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("BlockHeaderHashIndexKey.MarshalBinary: incorrect BlockHash length")
	}
	if len(b.Prefix) != 2 {
		return nil, errorz.ErrInvalid{}.New("BlockHeaderHashIndexKey.MarshalBinary: incorrect Prefix length")
	}
	key := []byte{}
	Prefix := utils.CopySlice(b.Prefix)
	BlockHash := utils.CopySlice(b.BlockHash)
	key = append(key, Prefix...)
	key = append(key, BlockHash...)
	return key, nil
}
