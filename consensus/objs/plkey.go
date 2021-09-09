package objs

import (
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
)

// PendingLeafKey ...
type PendingLeafKey struct {
	Prefix []byte
	Key    []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// PendingLeafKey object
func (b *PendingLeafKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if len(data) != 34 {
		return errorz.ErrInvalid{}.New("Invalid data for BlockHeaderHeightKey unmarshalling")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	b.Key = utils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the PendingLeafKey object and returns the canonical
// byte slice
func (b *PendingLeafKey) MarshalBinary() ([]byte, error) {
	if b == nil || len(b.Prefix) != 2 || len(b.Key) != 32 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	key = append(key, utils.CopySlice(b.Prefix)...)
	key = append(key, utils.CopySlice(b.Key)...)
	return key, nil
}
