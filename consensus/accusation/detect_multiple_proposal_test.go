package accusation

import (
	"strconv"
	"testing"

	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/crypto"
	"github.com/stretchr/testify/assert"
)

// TestNoMultipleProposalBehavior tests no bad behavior detected
func TestNoMultipleProposalBehavior(t *testing.T) {
	lrs := &lstate.RoundStates{}
	lrs.OwnState = &objs.OwnState{
		MaxBHSeen: &objs.BlockHeader{},
	}
	//txHashes := [][][]byte{}
	txhash := crypto.Hasher([]byte(strconv.Itoa(1)))
	txHshLst := [][]byte{txhash}
	txRoot, err := objs.MakeTxRoot(txHshLst)
	assert.Nil(t, err)
	//txHashes = append(txHashes, txHshLst)
	bclaims := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		TxCount:    1,
		PrevBlock:  crypto.Hasher([]byte("foo")),
		TxRoot:     txRoot,
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	lrs.OwnState.MaxBHSeen.BClaims = bclaims
	vSetMap := make(map[string]bool)
	vSetMap["aaaaa"] = true
	lrs.ValidatorSet = &objs.ValidatorSet{
		Validators: []*objs.Validator{
			{
				VAddr: []byte("aaaaa"),
			},
		},
		ValidatorVAddrSet: vSetMap,
	}
	bh, err := lrs.OwnState.MaxBHSeen.BlockHash()
	assert.Nil(t, err)

	// round state
	rclaims := &objs.RClaims{
		ChainID:   1,
		Height:    2,
		Round:     1,
		PrevBlock: bh,
	}
	rs := &objs.RoundState{}
	proposal0 := &objs.Proposal{
		Proposer: []byte("aaaaa"),
		PClaims: &objs.PClaims{
			RCert: &objs.RCert{
				RClaims: rclaims,
			},
		},
	}
	proposal1 := &objs.Proposal{
		PClaims: &objs.PClaims{
			RCert: &objs.RCert{
				RClaims: rclaims,
			},
		},
	}
	rs.Proposal = proposal0
	rs.ConflictingProposal = proposal1

	acc, found := detectMultipleProposal(rs, lrs)
	assert.False(t, found)
	assert.Nil(t, acc)
}

// TestDetectMultipleProposalBehavior tests malicious behavior detection
func TestDetectMultipleProposalBehavior(t *testing.T) {
	bnVal := &crypto.BNGroupValidator{}
	secpVal := &crypto.Secp256k1Validator{}
	bclaimsList, txHashListList, err := generateChain(2)
	if err != nil {
		t.Fatal(err)
	}
	bclaims := bclaimsList[0]
	bhsh, err := bclaims.BlockHash()
	if err != nil {
		t.Fatal(err)
	}
	gk := &crypto.BNGroupSigner{}
	err = gk.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	sig, err := gk.Sign(bhsh)
	if err != nil {
		t.Fatal(err)
	}
	secpSigner := &crypto.Secp256k1Signer{}
	err = secpSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	secpSigner2 := &crypto.Secp256k1Signer{}
	err = secpSigner2.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}
	bh := &objs.BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHashListList[0],
	}
	err = bh.ValidateSignatures(bnVal)
	if err != nil {
		t.Fatal(err)
	}
	rcert, err := bh.GetRCert()
	if err != nil {
		t.Fatal(err)
	}
	err = rcert.ValidateSignature(bnVal)
	if err != nil {
		t.Fatal(err)
	}
	pclms := &objs.PClaims{
		BClaims: bclaimsList[1],
		RCert:   rcert,
	}
	prop := &objs.Proposal{
		PClaims:  pclms,
		TxHshLst: txHashListList[1],
	}
	err = prop.Sign(secpSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = prop.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}
	propBytes, err := prop.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	prop2 := &objs.Proposal{}
	err = prop2.UnmarshalBinary(propBytes)
	if err != nil {
		t.Fatal(err)
	}
	// ensure prop.PClaims != prop2.PClaims
	prop2.PClaims.RCert.SigGroup = crypto.Hasher([]byte("blah"))
	assert.NotEqual(t, prop.PClaims.RCert.SigGroup, prop2.PClaims.RCert.SigGroup)
	// sign proposal2
	err = prop2.Sign(secpSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = prop2.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}

	rs := &objs.RoundState{}
	prop.Proposer = []byte("aaaaa")
	prop2.Proposer = []byte("aaaaa")
	rs.Proposal = prop
	rs.ConflictingProposal = prop2

	lrs := &lstate.RoundStates{}
	lrs.OwnState = &objs.OwnState{
		MaxBHSeen: &objs.BlockHeader{},
	}
	lrs.OwnState.MaxBHSeen.BClaims = bclaims
	vSetMap := make(map[string]bool)
	vSetMap["aaaaa"] = true
	lrs.ValidatorSet = &objs.ValidatorSet{
		Validators: []*objs.Validator{
			{
				VAddr: []byte("aaaaa"),
			},
		},
		ValidatorVAddrSet: vSetMap,
	}

	acc, found := detectMultipleProposal(rs, lrs)
	assert.True(t, found)
	assert.NotNil(t, acc)
}
