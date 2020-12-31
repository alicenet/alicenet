package objs

import (
	"github.com/MadBase/MadNet/errorz"
	gUtils "github.com/MadBase/MadNet/utils"
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
	if len(data) != 34 {
		return errorz.ErrInvalid{}.New("Invalid data for PendingNodeKey unmarshalling")
	}
	b.Prefix = gUtils.CopySlice(data[0:2])
	b.Key = gUtils.CopySlice(data[2:])
	return nil
}

// MarshalBinary takes the PendingNodeKey object and returns the canonical
// byte slice
func (b *PendingNodeKey) MarshalBinary() ([]byte, error) {
	if b == nil || len(b.Prefix) != 2 || len(b.Key) != 32 {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	key := []byte{}
	key = append(key, gUtils.CopySlice(b.Prefix)...)
	key = append(key, gUtils.CopySlice(b.Key)...)
	return key, nil
}
