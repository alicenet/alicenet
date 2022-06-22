package objs

import (
	"bytes"

	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// SafeToProceedKey ...
type SafeToProceedKey struct {
	Prefix []byte
	Height uint32
}

// MarshalBinary takes the SafeToProceedKey object and returns
// the canonical byte slice
func (b *SafeToProceedKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("SafeToProceedKey.MarshalBinary; stpk not initialized")
	}
	if b.Height == 0 {
		return nil, errorz.ErrInvalid{}.New("SafeToProceedKey.MarshalBinary; height is zero")
	}
	key := []byte{}
	Prefix := utils.CopySlice(b.Prefix)
	Height := utils.MarshalUint32(b.Height)
	key = append(key, Prefix...)
	key = append(key, []byte("|")...)
	key = append(key, Height...)
	key = append(key, []byte("|")...)
	return key, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// SafeToProceedKey object
func (b *SafeToProceedKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("SafeToProceedKey.UnmarshalBinary; stpk not initialized")
	}
	splitData := bytes.Split(data, []byte("|"))
	if len(splitData) != 3 {
		return errorz.ErrCorrupt
	}
	b.Prefix = splitData[0]
	Height, err := utils.UnmarshalUint32(splitData[1])
	if err != nil {
		return err
	}
	if Height == 0 {
		return errorz.ErrInvalid{}.New("SafeToProceedKey.UnmarshalBinary; height is zero")
	}
	if len(splitData[2]) != 0 {
		return errorz.ErrInvalid{}.New("SafeToProceedKey.UnmarshalBinary; invalid key")
	}
	b.Height = Height
	return nil
}
