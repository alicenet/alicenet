package objs

import (
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
)

// PendingHdrLeafKey ...
type PendingHdrLeafKey struct {
	Prefix []byte
	Key    []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// PendingHdrLeafKey object
func (b *PendingHdrLeafKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("PendingHdrLeafKey.UnmarshalBinary; phlk not initialized")
	}
	if len(data) != constants.HashLen+2 {
		return errorz.ErrInvalid{}.New("PendingHdrLeafKey.UnmarshalBinary; incorrect state length")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	b.Key = utils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the PendingHdrLeafKey object and returns the canonical
// byte slice
func (b *PendingHdrLeafKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("PendingHdrLeafKey.MarshalBinary; phlk not initialized")
	}
	if len(b.Prefix) != 2 {
		return nil, errorz.ErrInvalid{}.New("PendingHdrLeafKey.MarshalBinary; incorrect Prefix length")
	}
	if len(b.Key) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("PendingHdrLeafKey.MarshalBinary; incorrect key length")
	}
	key := []byte{}
	key = append(key, utils.CopySlice(b.Prefix)...)
	key = append(key, utils.CopySlice(b.Key)...)
	return key, nil
}
