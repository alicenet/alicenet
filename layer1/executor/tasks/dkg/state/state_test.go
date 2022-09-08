//go:build integration

package state_test

import (
	"bytes"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"

	"github.com/alicenet/alicenet/layer1/executor/tasks/dkg/state"
)

func TestDKGState_ParticipantCopy(t *testing.T) {
	p := &state.Participant{}
	addrBytes := make([]byte, 20)
	addrBytes[0] = 255
	addrBytes[19] = 255
	p.Address.SetBytes(addrBytes)
	publicKey := [2]*big.Int{big.NewInt(1), big.NewInt(2)}
	p.PublicKey = publicKey
	index := 13
	p.Index = index

	c := p.Copy()
	pBytes := p.Address.Bytes()
	cBytes := c.Address.Bytes()
	if !bytes.Equal(pBytes, cBytes) {
		t.Fatal("bytes do not match")
	}
	pPubKey := p.PublicKey
	cPubKey := c.PublicKey
	if pPubKey[0].Cmp(cPubKey[0]) != 0 || pPubKey[1].Cmp(cPubKey[1]) != 0 {
		t.Fatal("public keys do not match")
	}
	if p.Index != c.Index {
		t.Fatal("Indices do not match")
	}

	pString := p.String()
	cString := c.String()
	if pString != cString {
		t.Fatal("strings do not match")
	}
}

func TestDKGState_ParticipantListExtractIndices(t *testing.T) {
	p1 := &state.Participant{Index: 1}
	p2 := &state.Participant{Index: 2}
	p3 := &state.Participant{Index: 3}
	p4 := &state.Participant{Index: 4}

	pl := state.ParticipantList{p4, p2, p3, p1}
	indices := []int{4, 2, 3, 1}
	retIndices := pl.ExtractIndices()
	if len(indices) != len(retIndices) {
		t.Fatal("invalid indices")
	}
	for k := 0; k < len(indices); k++ {
		if indices[k] != retIndices[k] {
			t.Fatal("invalid indices when looping")
		}
	}
}

func TestDKGState_MarshalAndUnmarshalBigInt(t *testing.T) {
	// generate transport keys
	priv, pub, err := state.GenerateKeys()
	assert.Nil(t, err)

	// marshal privkey
	rawPrivData, err := json.Marshal(priv)
	assert.Nil(t, err)
	rawPubData, err := json.Marshal(pub)
	assert.Nil(t, err)

	priv2 := &big.Int{}
	pub2 := [2]*big.Int{}

	err = json.Unmarshal(rawPrivData, priv2)
	assert.Nil(t, err)
	err = json.Unmarshal(rawPubData, &pub2)
	assert.Nil(t, err)

	assert.Equal(t, priv, priv2)
	assert.Equal(t, pub, pub2)
}

func TestDKGState_MarshalAndUnmarshalAccount(t *testing.T) {
	addr := common.Address{}
	addr.SetBytes([]byte("546F99F244b7B58B855330AE0E2BC1b30b41302F"))

	// create a DkgState obj
	acct := accounts.Account{
		Address: addr,
		URL: accounts.URL{
			Scheme: "http",
			Path:   "",
		},
	}

	// marshal acct
	rawData, err := json.Marshal(acct)
	assert.Nil(t, err)

	acct2 := &accounts.Account{}

	err = json.Unmarshal(rawData, acct2)
	assert.Nil(t, err)

	assert.Equal(t, acct, *acct2)
}

func TestDKGState_MarshalAndUnmarshalParticipant(t *testing.T) {
	addr := common.Address{}
	addr.SetBytes([]byte("546F99F244b7B58B855330AE0E2BC1b30b41302F"))

	// generate transport keys
	_, pub, err := state.GenerateKeys()
	assert.Nil(t, err)

	// create a state.Participant obj
	participant := state.Participant{
		Address:   addr,
		Index:     1,
		PublicKey: pub,
		Nonce:     1,
		Phase:     state.RegistrationOpen,
	}

	// marshal
	rawData, err := json.Marshal(participant)
	assert.Nil(t, err)

	t.Logf("rawData: %s", rawData)

	participant2 := &state.Participant{}

	err = json.Unmarshal(rawData, participant2)
	assert.Nil(t, err)
	assert.Equal(t, participant.PublicKey, participant2.PublicKey)
}

func TestDKGState_MarshalAndUnmarshalDkgState(t *testing.T) {
	addr := common.Address{}
	addr.SetBytes([]byte("546F99F244b7B58B855330AE0E2BC1b30b41302F"))

	// create a DkgState obj
	state1 := state.NewDkgState(accounts.Account{
		Address: addr,
		URL: accounts.URL{
			Scheme: "file",
			Path:   "",
		},
	})

	// generate transport keys
	priv, pub, err := state.GenerateKeys()
	assert.Nil(t, err)
	state1.TransportPrivateKey = priv
	state1.TransportPublicKey = pub

	// marshal
	rawData, err := json.Marshal(state1)
	assert.Nil(t, err)

	t.Logf("rawData: %s", rawData)

	state2 := &state.DkgState{}

	err = json.Unmarshal(rawData, state2)
	assert.Nil(t, err)
	assert.Equal(t, state1.TransportPrivateKey, state2.TransportPrivateKey)
	assert.Equal(t, state1.TransportPublicKey, state2.TransportPublicKey)
}
