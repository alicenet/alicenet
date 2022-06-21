package brontide

import (
	"encoding/binary"
	"net"

	"github.com/alicenet/alicenet/crypto/secp256k1"
)

type NetAddress struct {
	IdentityKey *secp256k1.PublicKey
	Address     net.Addr
}

// Serializes a chain identifier into bytes.
func marshalUint32(c uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:], c)
	return b
}

// Converts serialized bytes back into a ChainIdentifier.
func unmarshalUint32(b [4]byte) uint32 {
	c := binary.BigEndian.Uint32(b[:])
	return c
}
