package objs

import (
	"github.com/MadBase/MadNet/constants"
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
		return errorz.ErrInvalid{}.New("PendingLeafKey.UnmarshalBinary; plk not initialized")
	}
	if len(data) != constants.HashLen+2 {
		return errorz.ErrInvalid{}.New("PendingLeafKey.UnmarshalBinary; incorrect state length")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	b.Key = utils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the PendingLeafKey object and returns the canonical
// byte slice
func (b *PendingLeafKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("PendingLeafKey.MarshalBinary; plk not initialized")
	}
	if len(b.Prefix) != 2 {
		return nil, errorz.ErrInvalid{}.New("PendingLeafKey.MarshalBinary; incorrect Prefix length")
	}
	if len(b.Key) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("PendingLeafKey.MarshalBinary; incorrect key length")
	}
	key := []byte{}
	key = append(key, utils.CopySlice(b.Prefix)...)
	key = append(key, utils.CopySlice(b.Key)...)
	return key, nil
}
