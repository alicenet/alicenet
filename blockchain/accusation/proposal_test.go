package accusation_test

import (
	"testing"

	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/crypto"
	gec "github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
)

func TestDoubleProposal(t *testing.T) {

	voo := &objs.RoundState{
		VAddr: []byte{},
		Proposal: &objs.Proposal{
			Signature: []byte{},
			PClaims: &objs.PClaims{
				RCert: &objs.RCert{
					SigGroup: []byte{},
					RClaims: &objs.RClaims{
						ChainID:   42,
						Height:    73,
						Round:     2,
						PrevBlock: []byte{},
					},
				},
				BClaims: &objs.BClaims{
					ChainID:   42,
					Height:    73,
					PrevBlock: []byte{},
				},
			},
			Proposer: []byte{},
		},
	}

	signer := &crypto.Secp256k1Signer{}
	err := signer.SetPrivk(crypto.Hasher([]byte("secret")))
	assert.Nil(t, err)

	pclaimsRaw, err := voo.Proposal.PClaims.MarshalBinary()
	assert.Nil(t, err)

	signature, err := signer.Sign(pclaimsRaw)
	assert.Nil(t, err)

	t.Logf("pclaimsRaw: 0x%x", pclaimsRaw)
	t.Logf("signature: 0x%x", signature)
}

func TestSignature(t *testing.T) {

	d := crypto.Hasher([]byte("secret"))
	privateKey, err := gec.ToECDSA(d)
	assert.Nil(t, err)
	t.Logf("private key: %x", d)

	address := gec.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("public key: %x", gec.FromECDSAPub(&privateKey.PublicKey))
	t.Logf("address: %v", address.Hex())

	message := []byte("The quick brown fox did something")
	t.Logf("message: %x", message)
	hashedMessage := gec.Keccak256(message)
	t.Logf("hashMessage: %x", hashedMessage)

	sig, err := gec.Sign(hashedMessage, privateKey)
	assert.Nil(t, err)
	t.Logf("sig: %x", sig)
}

func TestSignatureOther(t *testing.T) {

	d := crypto.Hasher([]byte("secret"))
	signer := &crypto.Secp256k1Signer{}
	err := signer.SetPrivk(d)
	assert.Nil(t, err)
	t.Logf("private key: %x", d)

	publicKey, err := signer.Pubkey()
	assert.Nil(t, err)

	t.Logf("public key: %x", publicKey)

	address := gec.Keccak256(publicKey[1:])[12:]
	assert.Equal(t, 20, len(address))

	t.Logf("address: %x", address)

	message := []byte("The quick brown fox did something")
	t.Logf("message: %x", message)

	sig, err := signer.Sign(message)
	assert.Nil(t, err)

	t.Logf("sig: %x", sig)
}
