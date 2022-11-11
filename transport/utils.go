package transport

import (
	"encoding/binary"
	"encoding/hex"

	secp256k1_crypto "github.com/alicenet/alicenet/crypto/secp256k1"
	secp256k1_v4 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/lightningnetwork/lnd/keychain"
)

// Converts a uint32 into a hex string.
func uint32ToHexString(v uint32) string {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b[:], v)
	return hex.EncodeToString(b)
}

// Converts a hex string into a uint32.
// If the string will overflow a uint32,
// an error is raised.
func uint32FromHexString(v string) (uint32, error) {
	hb, err := hex.DecodeString(v)
	if err != nil {
		return 0, err
	}
	if len(hb) > 4 {
		return 0, ErrUint32Overflow
	}
	return binary.BigEndian.Uint32(hb[:]), nil
}

func convertCryptoSecpPubKey2DecredSecpPubKey(pubKey *secp256k1_crypto.PublicKey) (*secp256k1_v4.PublicKey, error) {
	newPubKey, err := secp256k1_v4.ParsePubKey(pubKey.SerializeCompressed())
	if err != nil {
		return nil, err
	}

	return newPubKey, nil
}

func convertDecredSecpPubKey2CryptoSecpPubKey(pubKey *secp256k1_v4.PublicKey) (*secp256k1_crypto.PublicKey, error) {

	newPubKey, err := secp256k1_crypto.ParsePubKey(pubKey.SerializeCompressed(), secp256k1_crypto.S256())
	if err != nil {
		return nil, err
	}

	return newPubKey, nil
}

func convertCryptoPrivateKey2KeychainPrivateKey(privKey *secp256k1_crypto.PrivateKey) keychain.SingleKeyECDH {
	privateKey := secp256k1_v4.PrivKeyFromBytes(privKey.Serialize())
	return &keychain.PrivKeyECDH{PrivKey: privateKey}
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
