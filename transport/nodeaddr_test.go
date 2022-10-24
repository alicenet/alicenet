package transport

import (
	"github.com/alicenet/alicenet/config"
	"github.com/alicenet/alicenet/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNodeAddr(t *testing.T) {
	chaindID := 4444
	config.Configuration.Chain.ID = chaindID
	tkey, err := newTransportPrivateKey()
	if err != nil {
		t.Fatal(err)
	}
	pubk := publicKeyFromPrivateKey(tkey)
	addr := &NodeAddr{
		host:     "127.0.0.1",
		port:     3000,
		identity: pubk,
		chainID:  types.ChainIdentifier(chaindID),
	}
	p2pAddr := addr.P2PAddr()
	addr2if, err := NewNodeAddr(p2pAddr)
	if err != nil {
		t.Fatal(err)
	}
	addr2 := addr2if.(*NodeAddr)
	assert.Equal(t, addr, addr2)

	addr3if, err := (*NodeAddr).Unmarshal(nil, p2pAddr)
	if err != nil {
		t.Fatal(err)
	}
	addr3 := addr3if.(*NodeAddr)
	assert.Equal(t, addr, addr3)
	t.Log(addr3.P2PAddr())
}
