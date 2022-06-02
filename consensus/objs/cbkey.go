package objs

import (
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
)

// BlockHeaderHeightKey ...
type BlockHeaderHeightKey struct {
	Prefix []byte
	Height uint32
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// BlockHeaderHeightKey object
func (b *BlockHeaderHeightKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("BlockHeaderHeightKey.UnmarshalBinary; bhhk not initialized")
	}
	if len(data) != 6 {
		return errorz.ErrInvalid{}.New("BlockHeaderHeightKey.UnmarshalBinary; incorrect state length")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	nb, err := utils.UnmarshalUint32(data[2:6])
	if err != nil {
		return err
	}
	if nb == 0 {
		return errorz.ErrInvalid{}.New("BlockHeaderHeightKey.UnmarshalBinary; height is zero")
	}
	b.Height = nb
	return nil
}

// MarshalBinary takes the BlockHeaderHeightKey object and returns the canonical
// byte slice
func (b *BlockHeaderHeightKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("BlockHeaderHeightKey.MarshalBinary; bhhk not initialized")
	}
	if len(b.Prefix) != 2 {
		return nil, errorz.ErrInvalid{}.New("BlockHeaderHeightKey.MarshalBinary; incorrect Prefix length")
	}
	if b.Height == 0 {
		return nil, errorz.ErrInvalid{}.New("BlockHeaderHeightKey.MarshalBinary; height is zero")
	}
	key := []byte{}
	Prefix := utils.CopySlice(b.Prefix)
	nb := utils.MarshalUint32(b.Height)
	key = append(key, Prefix...)
	key = append(key, nb...)
	return key, nil
}
