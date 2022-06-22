package objs

import (
	"bytes"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// TxCacheKey ...
type TxCacheKey struct {
	Prefix []byte
	Height uint32
	TxHash []byte
}

// MarshalBinary takes the TxCacheKey object and returns
// the canonical byte slice
func (b *TxCacheKey) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("TxCacheKey.MarshalBinary; tck not initialized")
	}
	if b.Height == 0 {
		return nil, errorz.ErrInvalid{}.New("TxCacheKey.MarshalBinary; height is zero")
	}
	if len(b.TxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("TxCacheKey.MarshalBinary; incorrect txhash length")
	}
	key := []byte{}
	Prefix := utils.CopySlice(b.Prefix)
	TxHash := utils.CopySlice(b.TxHash)
	Height := utils.MarshalUint32(b.Height)
	key = append(key, Prefix...)
	key = append(key, []byte("|")...)
	key = append(key, Height...)
	key = append(key, []byte("|")...)
	key = append(key, TxHash...)
	return key, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// TxCacheKey object
func (b *TxCacheKey) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("TxCacheKey.UnmarshalBinary; tck not initialized")
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
		return errorz.ErrInvalid{}.New("TxCacheKey.UnmarshalBinary; height is zero")
	}
	b.Height = Height
	TxHash := utils.CopySlice(splitData[2])
	if len(TxHash) != constants.HashLen {
		return errorz.ErrInvalid{}.New("TxCacheKey.UnmarshalBinary; incorrect txhash length")
	}
	b.TxHash = TxHash
	return nil
}

// MakeIterKey ...
func (b *TxCacheKey) MakeIterKey() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("TxCacheKey.MakeIterKey; tck not initialized")
	}
	if b.Height == 0 {
		return nil, errorz.ErrInvalid{}.New("TxCacheKey.MakeIterKey; height is zero")
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
