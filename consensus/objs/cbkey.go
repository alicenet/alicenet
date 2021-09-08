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
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if len(data) != 6 {
		return errorz.ErrInvalid{}.New("Invalid data for BlockHeaderHeightKey unmarshalling")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	nb, err := utils.UnmarshalUint32(data[2:6])
	if err != nil {
		return err
	}
	if nb == 0 {
		return errorz.ErrInvalid{}.New("Invalid data for BlockHeaderHeightKey unmarshalling; height == 0")
	}
	b.Height = nb
	return nil
}

// MarshalBinary takes the BlockHeaderHeightKey object and returns the canonical
// byte slice
func (b *BlockHeaderHeightKey) MarshalBinary() ([]byte, error) {
	if b == nil || len(b.Prefix) != 2 || b.Height == 0 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	Prefix := utils.CopySlice(b.Prefix)
	nb := utils.MarshalUint32(b.Height)
	key = append(key, Prefix...)
	key = append(key, nb...)
	return key, nil
}
