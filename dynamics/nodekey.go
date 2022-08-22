package dynamics

import (
	"bytes"

	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/utils"
)

// NodeKey stores the necessary information to load a Node
type NodeKey struct {
	prefix []byte
	epoch  uint32
}

func makeNodeKey(epoch uint32) (*NodeKey, error) {
	if epoch == 0 {
		return nil, ErrZeroEpoch
	}
	nk := &NodeKey{
		prefix: dbprefix.PrefixStorageNodeKey(),
		epoch:  epoch,
	}
	return nk, nil
}

// Marshal converts NodeKey into the byte slice
func (nk *NodeKey) Marshal() ([]byte, error) {
	if !bytes.Equal(nk.prefix, dbprefix.PrefixStorageNodeKey()) {
		return nil, ErrInvalidNodeKey
	}
	epochBytes := utils.MarshalUint32(nk.epoch)
	key := []byte{}
	key = append(key, nk.prefix...)
	key = append(key, epochBytes...)
	return key, nil
}
