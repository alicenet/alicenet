package transport

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrivateKey(t *testing.T) {
	_, err := NewTransportPrivateKey()
	if err != nil {
		t.Error("Failed to create private key.")
	}
	nodeKey, err := newTransportPrivateKey()
	if err != nil {
		t.Error("Failed to create private key.")
	}
	nodeKeyHex := serializeTransportPrivateKey(nodeKey)
	nodeKeyAfter, err := deserializeTransportPrivateKey(nodeKeyHex)
	if err != nil {
		t.Error("Deserialization failed for private key.")
	}
	n1b := nodeKey.Serialize()
	n2b := nodeKeyAfter.Serialize()
	assert.True(t, bytes.Equal(n1b, n2b), "PrivateKey serialization failed.")
	_, err = deserializeTransportPrivateKey("")
	if err == nil {
		t.Error("Deserialization did not fail as expected for private key.")
	}
}
