package objs

import (
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
)

// PendingNodeKey ...
type PendingNodeKey struct {
	Prefix []byte
	Key    []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// PendingNodeKey object
func (b *PendingNodeKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if len(data) != constants.HashLen+2 {
		return errorz.ErrInvalid{}.New("Invalid data for PendingNodeKey unmarshalling")
	}
	b.Prefix = utils.CopySlice(data[0:2])
	b.Key = utils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the PendingNodeKey object and returns the canonical
// byte slice
func (b *PendingNodeKey) MarshalBinary() ([]byte, error) {
	if b == nil || len(b.Prefix) != 2 || len(b.Key) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	key = append(key, utils.CopySlice(b.Prefix)...)
	key = append(key, utils.CopySlice(b.Key)...)
	return key, nil
}
