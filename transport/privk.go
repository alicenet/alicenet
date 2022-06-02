package transport

import (
	"encoding/hex"

	"github.com/MadBase/MadNet/crypto/secp256k1"
)

// NewTransportPrivateKey returns a new transport private key as a hex string.
// This key is used for the creation of an authenticated and encrypted stream
// of state between peers.
func NewTransportPrivateKey() (string, error) {
	privateKey, err := newTransportPrivateKey()
	if err != nil {
		return "", nil
	}
	return serializeTransportPrivateKey(privateKey), nil
}

// Remaining functions broken out for testing purposes

func newTransportPrivateKey() (*secp256k1.PrivateKey, error) {
	return secp256k1.NewPrivateKey(secp256k1.S256())
}

func serializeTransportPrivateKey(privateKey *secp256k1.PrivateKey) string {
	privateKeyBytes := privateKey.Serialize()
	if len(privateKeyBytes) < 16 {
		panic("Invalid private key.")
	}
	return hex.EncodeToString(privateKeyBytes)
}

func deserializeTransportPrivateKey(privateKeyHex string) (*secp256k1.PrivateKey, error) {
	if len(privateKeyHex) == 0 {
		return nil, ErrEmptyPrivKey
	}
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		return nil, err
	}
	if len(privateKeyBytes) < 16 {
		return nil, ErrInvalidPrivKey
	}
	privateKey, _ := secp256k1.PrivKeyFromBytes(secp256k1.S256(), privateKeyBytes)
	return privateKey, nil
}

func publicKeyFromPrivateKey(privateKey *secp256k1.PrivateKey) *secp256k1.PublicKey {
	return privateKey.PubKey()
}
