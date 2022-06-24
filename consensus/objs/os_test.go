package objs

import (
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/crypto"
	bn256 "github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/stretchr/testify/assert"
)

func TestOwnState(t *testing.T) {
	//Vaddr
	secret1 := big.NewInt(100)
	secret2 := big.NewInt(101)
	secret3 := big.NewInt(102)
	secret4 := big.NewInt(103)

	big1 := big.NewInt(1)
	big2 := big.NewInt(2)

	privCoefs1 := []*big.Int{secret1, big1, big2}
	privCoefs2 := []*big.Int{secret2, big1, big2}
	privCoefs3 := []*big.Int{secret3, big1, big2}
	privCoefs4 := []*big.Int{secret4, big1, big2}

	share1to1 := bn256.PrivatePolyEval(privCoefs1, 1)
	share2to1 := bn256.PrivatePolyEval(privCoefs2, 1)
	share3to1 := bn256.PrivatePolyEval(privCoefs3, 1)
	share4to1 := bn256.PrivatePolyEval(privCoefs4, 1)

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := bn256.GenerateGroupSecretKeyPortion(listOfSS1)
	gpk1 := new(bn256.G2).ScalarBaseMult(gsk1)

	secpSigner := &crypto.Secp256k1Signer{}
	err := secpSigner.SetPrivk(crypto.Hasher(gpk1.Marshal()))
	if err != nil {
		panic(err)
	}
	secpKey, err := secpSigner.Pubkey()
	if err != nil {
		panic(err)
	}

	//BlockHeader
	bclaimsList, txHashListList, err := generateChain(1)
	if err != nil {
		t.Fatal(err)
	}
	bclaims := bclaimsList[0]
	bhsh, err := bclaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	gk := crypto.BNGroupSigner{}
	err = gk.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	sig, err := gk.Sign(bhsh)
	if err != nil {
		t.Fatal(err)
	}
	bh := &BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHashListList[0],
	}

	ows := make([]*OwnState, 1)
	binary, err := ows[0].MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error")
	}

	// os := &OwnState{}
	// os = nil
	// binary, err = os.MarshalBinary()
	// log.Println(err)
	// if err == nil {
	// 	t.Fatal("Should raise an error")
	// }

	err = ows[0].UnmarshalBinary(binary)
	if err == nil {
		t.Fatal("Should raise an error")
	}

	ows[0] = &OwnState{
		VAddr:             secpKey,
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}

	binary, err = ows[0].MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	err = ows[0].UnmarshalBinary(binary)
	if err != nil {
		t.Fatal(err)
	}

	newOs, err := ows[0].Copy()
	if err != nil {
		t.Fatal(err)
	}

	isSync := newOs.IsSync()
	assert.True(t, isSync)
}
