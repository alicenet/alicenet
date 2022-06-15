package lstate

import (
	"context"
	"math/big"
	"strconv"
	"testing"

	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	bn256 "github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func TestStore_LoadLocalState_Ok(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		result, err := store.LoadLocalState(txn)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		assert.NotNil(t, result)
		return nil
	})
}

func TestStore_LoadLocalState_Errors(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)

	_ = store.database.Update(func(txn *badger.Txn) error {
		result, err := store.LoadLocalState(txn)
		if err == nil {
			t.Fatal("Should have raised error (1)")
		}
		assert.Nil(t, result)

		err = store.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		result, err = store.LoadLocalState(txn)
		if err == nil {
			t.Fatal("Should have raised error (2)")
		}
		assert.Nil(t, result)

		err = store.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}
		result, err = store.LoadLocalState(txn)
		if err == nil {
			t.Fatal("Should have raised error (3)")
		}
		assert.Nil(t, result)
		return nil
	})
}

func TestStore_WriteState_Ok(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(os, rs, vs, &objs.OwnValidatingState{})

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStore_WriteState_Errors(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	rss := createRoundStates(&objs.OwnState{}, rs, vs, &objs.OwnValidatingState{})

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.WriteState(txn, rss)
		if err == nil {
			t.Fatal("Should have raised error (1)")
		}

		rss = createRoundStates(os, rs, vs, &objs.OwnValidatingState{})
		rss.PeerStateMap[string(os.VAddr)] = &objs.RoundState{VAddr: make([]byte, constants.HashLen)}
		err = store.WriteState(txn, rss)
		if err == nil {
			t.Fatal("Should have raised error (2)")
		}

		return nil
	})
}

func TestStore_WriteState_ConflictLockedValue(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	ovs := &objs.OwnValidatingState{
		RoundStarted:         1,
		PreVoteStepStarted:   1,
		PreCommitStepStarted: 1,
	}
	_, lv := createProposal(t)
	lv.PClaims.RCert.RClaims.Height = 1
	ovs.LockedValue = lv

	rss := createRoundStates(os, rs, vs, ovs)

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStore_WriteState_ConflictValidValue(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)
	ovs := &objs.OwnValidatingState{
		RoundStarted:         1,
		PreVoteStepStarted:   1,
		PreCommitStepStarted: 1,
	}
	_, vv := createProposal(t)
	vv.PClaims.RCert.RClaims.Height = 1
	ovs.ValidValue = vv

	rss := createRoundStates(os, rs, vs, ovs)

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.WriteState(txn, rss)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		return nil
	})
}

func TestStore_GetDropData_Ok_OS_Validator(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		isValidator, isSync, chainID, height, round, err := store.GetDropData(txn)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		assert.Equal(t, uint32(0x1), chainID)
		assert.Equal(t, uint32(0x1), height)
		assert.Equal(t, uint32(0x1), round)
		assert.True(t, isValidator)
		assert.True(t, isSync)

		return nil
	})
}

func TestStore_GetDropData_Ok_OS_NotValidator(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, nil, rs)

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		isValidator, isSync, chainID, height, round, err := store.GetDropData(txn)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		assert.Equal(t, uint32(0x1), chainID)
		assert.Equal(t, uint32(0x1), height)
		assert.Equal(t, uint32(0x1), round)
		assert.False(t, isValidator)
		assert.True(t, isSync)

		return nil
	})
}

//Proposal, PreVote, PreCommit, NextHeight
func TestStore_GetGossipValues_Ok1(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, pvl, _, pcl, _, _, nhl, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetBroadcastProposal(txn, pl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetBroadcastPreVote(txn, pvl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetBroadcastPreCommit(txn, pcl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetBroadcastNextHeight(txn, nhl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		p, pv, pvn, pc, pcn, nr, nh, err := store.GetGossipValues(txn)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		assert.NotNil(t, p)
		assert.NotNil(t, pv)
		assert.Nil(t, pvn)
		assert.NotNil(t, pc)
		assert.Nil(t, pcn)
		assert.Nil(t, nr)
		assert.NotNil(t, nh)

		return nil
	})
}

//Proposal, PreVoteNil, PreCommitNil, NextRound
func TestStore_GetGossipValues_Ok2(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	_, bnSigners, bnShares, secpSigners, secpPubks := makeSigners(t)
	if len(secpPubks) != len(bnShares) {
		t.Fatal("key length mismatch")
	}

	height := uint32(2)
	round := uint32(1)
	prevBlock := crypto.Hasher([]byte("0"))
	_, pl, _, pvnl, _, pcnl, nrl, _, _ := buildRound(t, bnSigners, bnShares, secpSigners, height, round, prevBlock)

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetBroadcastProposal(txn, pl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetBroadcastPreVoteNil(txn, pvnl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetBroadcastPreCommitNil(txn, pcnl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = store.database.SetBroadcastNextRound(txn, nrl[0])
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		p, pv, pvn, pc, pcn, nr, nh, err := store.GetGossipValues(txn)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		assert.NotNil(t, p)
		assert.Nil(t, pv)
		assert.NotNil(t, pvn)
		assert.Nil(t, pc)
		assert.NotNil(t, pcn)
		assert.NotNil(t, nr)
		assert.Nil(t, nh)

		return nil
	})
}

func TestStore_GetSyncToBH_Error(t *testing.T) {
	store := initStore(t)

	_ = store.database.Update(func(txn *badger.Txn) error {
		result, err := store.GetSyncToBH(txn)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		assert.Nil(t, result)
		return nil
	})
}

func TestStore_GetMaxBH_Error(t *testing.T) {
	store := initStore(t)

	_ = store.database.Update(func(txn *badger.Txn) error {
		result, err := store.GetMaxBH(txn)
		if err == nil {
			t.Fatalf("Should have raised error")
		}

		assert.Nil(t, result)
		return nil
	})
}

func TestStore_IsSync_Ok(t *testing.T) {
	store := initStore(t)
	os := createOwnState(t, 1)

	_ = store.database.Update(func(txn *badger.Txn) error {
		err := store.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		isSync, err := store.IsSync(txn)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		assert.True(t, isSync)
		return nil
	})
}

func initStore(t *testing.T) *Store {
	rawDb, err := utils.OpenBadger(context.Background().Done(), "", true)
	assert.Nil(t, err)
	database := &db.Database{}
	database.Init(rawDb)

	store := &Store{}
	store.Init(database)

	return store
}

func createProposal(t *testing.T) (*crypto.Secp256k1Signer, *objs.Proposal) {
	bclaimsList, bh := createBlockHeader(t, 1)
	rcert, err := bh.GetRCert()
	if err != nil {
		t.Fatal(err)
	}
	pclms := &objs.PClaims{
		BClaims: bclaimsList[0],
		RCert:   rcert,
	}
	pclmsBytes, err := pclms.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	secpSigner := &crypto.Secp256k1Signer{}
	err = secpSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	sig, err := secpSigner.Sign(pclmsBytes)
	if err != nil {
		t.Fatal(err)
	}
	prop := &objs.Proposal{
		PClaims:   pclms,
		Signature: sig,
		TxHshLst:  bh.TxHshLst,
	}

	err = prop.Sign(secpSigner)
	if err != nil {
		t.Fatal(err)
	}

	secpVal := &crypto.Secp256k1Validator{}
	bnVal := &crypto.BNGroupValidator{}
	err = prop.ValidateSignatures(secpVal, bnVal)
	if err != nil {
		t.Fatal(err)
	}

	return secpSigner, prop
}
func createRoundStates(os *objs.OwnState, rs *objs.RoundState, vs *objs.ValidatorSet, ovs *objs.OwnValidatingState) *RoundStates {
	rss := &RoundStates{
		height:             1,
		round:              1,
		OwnState:           os,
		ValidatorSet:       vs,
		OwnValidatingState: ovs,
		PeerStateMap:       make(map[string]*objs.RoundState),
	}

	rss.PeerStateMap[string(os.VAddr)] = rs

	for _, val := range vs.Validators {
		rss.PeerStateMap[string(val.VAddr)] = rs
	}

	rss.PeerStateMap[string(os.VAddr)] = rs
	return rss
}

func createValidatorsSet(t *testing.T, os *objs.OwnState, rs *objs.RoundState) *objs.ValidatorSet {
	vldtrs := []objects.Validator{
		createValidator("0x1", 1),
		createValidator("0x2", 2),
		createValidator("0x3", 3),
		createValidator("0x4", 4)}

	validators := make([]*objs.Validator, 0)

	for i, v := range vldtrs {
		g := &bn256.G2{}
		g.ScalarBaseMult(big.NewInt(int64(i)))
		ret := g.Marshal()
		val := &objs.Validator{
			VAddr:      v.Account.Bytes(),
			GroupShare: ret}

		validators = append(validators, val)
	}

	notBefore := uint32(1)
	vSet := &objs.ValidatorSet{
		Validators: validators,
		GroupKey:   rs.GroupKey,
		NotBefore:  notBefore,
	}

	if os != nil {
		g := &bn256.G2{}
		g.ScalarBaseMult(big.NewInt(int64(len(vSet.Validators))))
		ret := g.Marshal()
		osValidator := &objs.Validator{
			VAddr:      os.VAddr,
			GroupShare: ret,
		}
		vSet.Validators = append(vSet.Validators, osValidator)
	}

	vSet.ValidatorVAddrMap = make(map[string]int)
	vSet.ValidatorVAddrSet = make(map[string]bool)
	vSet.ValidatorGroupShareMap = make(map[string]int)
	vSet.ValidatorGroupShareSet = make(map[string]bool)
	for idx, v := range vSet.Validators {
		vSet.ValidatorVAddrMap[string(v.VAddr)] = idx
		vSet.ValidatorVAddrSet[string(v.VAddr)] = true
		vSet.ValidatorGroupShareMap[string(v.GroupShare)] = idx
		vSet.ValidatorGroupShareSet[string(v.GroupShare)] = true
	}

	return vSet
}

func createSharedKey(addr common.Address) [4]*big.Int {

	b := addr.Bytes()

	return [4]*big.Int{
		(&big.Int{}).SetBytes(b),
		(&big.Int{}).SetBytes(b),
		(&big.Int{}).SetBytes(b),
		(&big.Int{}).SetBytes(b)}
}

func createValidator(addrHex string, idx uint8) objects.Validator {
	addr := common.HexToAddress(addrHex)
	return objects.Validator{
		Account:   addr,
		Index:     idx,
		SharedKey: createSharedKey(addr),
	}
}

func createRoundState(t *testing.T, os *objs.OwnState) *objs.RoundState {
	groupSigner := &crypto.BNGroupSigner{}
	err := groupSigner.SetPrivk(crypto.Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	groupKey, _ := groupSigner.PubkeyShare()

	bnSigner := &crypto.BNGroupSigner{}
	err = bnSigner.SetPrivk(crypto.Hasher([]byte("secret2")))
	if err != nil {
		t.Fatal(err)
	}
	bnKey, _ := groupSigner.PubkeyShare()

	secpSigner := &crypto.Secp256k1Signer{}
	err = secpSigner.SetPrivk(crypto.Hasher([]byte("secret3")))
	if err != nil {
		t.Fatal(err)
	}

	prevBlock := make([]byte, constants.HashLen)
	sig, err := groupSigner.Sign(prevBlock)
	if err != nil {
		t.Fatal(err)
	}

	rs := &objs.RoundState{
		VAddr:      os.VAddr, // change done
		GroupKey:   groupKey,
		GroupShare: bnKey,
		GroupIdx:   127,
		RCert: &objs.RCert{
			SigGroup: sig,
			RClaims: &objs.RClaims{
				ChainID:   1,
				Height:    1,
				PrevBlock: prevBlock,
				Round:     1,
			},
		},
	}

	return rs
}

func createOwnState(t *testing.T, length int) *objs.OwnState {
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
	_, bh := createBlockHeader(t, length)

	return &objs.OwnState{
		VAddr:             secpKey,
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
}

func createBlockHeader(t *testing.T, length int) ([]*objs.BClaims, *objs.BlockHeader) {
	bclaimsList, txHashListList, err := generateChain(length)
	if err != nil {
		t.Fatal(err)
	}
	bclaims := bclaimsList[len(bclaimsList)-1]
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

	bh := &objs.BlockHeader{
		BClaims:  bclaims,
		SigGroup: sig,
		TxHshLst: txHashListList[0],
	}

	return bclaimsList, bh
}

func generateChain(length int) ([]*objs.BClaims, [][][]byte, error) {
	chain := []*objs.BClaims{}
	txHashes := [][][]byte{}
	txhash := crypto.Hasher([]byte(strconv.Itoa(1)))
	txHshLst := [][]byte{txhash}
	txRoot, err := objs.MakeTxRoot(txHshLst)
	if err != nil {
		return nil, nil, err
	}
	txHashes = append(txHashes, txHshLst)
	bclaims := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		TxCount:    1,
		PrevBlock:  crypto.Hasher([]byte("foo")),
		TxRoot:     txRoot,
		StateRoot:  crypto.Hasher([]byte("")),
		HeaderRoot: crypto.Hasher([]byte("")),
	}
	chain = append(chain, bclaims)
	for i := 1; i < length; i++ {
		bhsh, err := chain[i-1].BlockHash()
		if err != nil {
			return nil, nil, err
		}
		txhash := crypto.Hasher([]byte(strconv.Itoa(i)))
		txHshLst := [][]byte{txhash}
		txRoot, err := objs.MakeTxRoot(txHshLst)
		if err != nil {
			return nil, nil, err
		}
		txHashes = append(txHashes, txHshLst)
		bclaims := &objs.BClaims{
			ChainID:    1,
			Height:     uint32(len(chain) + 1),
			TxCount:    1,
			PrevBlock:  bhsh,
			TxRoot:     txRoot,
			StateRoot:  chain[i-1].StateRoot,
			HeaderRoot: chain[i-1].HeaderRoot,
		}
		chain = append(chain, bclaims)
	}
	return chain, txHashes, nil
}

func makeSigners(t *testing.T) ([]byte, []*crypto.BNGroupSigner, [][]byte, []*crypto.Secp256k1Signer, [][]byte) {
	s := new(crypto.BNGroupSigner)
	msg := []byte("A message to sign")

	secret1 := big.NewInt(100)
	secret2 := big.NewInt(101)
	secret3 := big.NewInt(102)
	secret4 := big.NewInt(103)

	msk := big.NewInt(0)
	msk.Add(msk, secret1)
	msk.Add(msk, secret2)
	msk.Add(msk, secret3)
	msk.Add(msk, secret4)
	msk.Mod(msk, bn256.Order)
	mpk := new(bn256.G2).ScalarBaseMult(msk)

	big1 := big.NewInt(1)
	big2 := big.NewInt(2)

	privCoefs1 := []*big.Int{secret1, big1, big2}
	privCoefs2 := []*big.Int{secret2, big1, big2}
	privCoefs3 := []*big.Int{secret3, big1, big2}
	privCoefs4 := []*big.Int{secret4, big1, big2}

	share1to1 := bn256.PrivatePolyEval(privCoefs1, 1)
	share1to2 := bn256.PrivatePolyEval(privCoefs1, 2)
	share1to3 := bn256.PrivatePolyEval(privCoefs1, 3)
	share1to4 := bn256.PrivatePolyEval(privCoefs1, 4)
	share2to1 := bn256.PrivatePolyEval(privCoefs2, 1)
	share2to2 := bn256.PrivatePolyEval(privCoefs2, 2)
	share2to3 := bn256.PrivatePolyEval(privCoefs2, 3)
	share2to4 := bn256.PrivatePolyEval(privCoefs2, 4)
	share3to1 := bn256.PrivatePolyEval(privCoefs3, 1)
	share3to2 := bn256.PrivatePolyEval(privCoefs3, 2)
	share3to3 := bn256.PrivatePolyEval(privCoefs3, 3)
	share3to4 := bn256.PrivatePolyEval(privCoefs3, 4)
	share4to1 := bn256.PrivatePolyEval(privCoefs4, 1)
	share4to2 := bn256.PrivatePolyEval(privCoefs4, 2)
	share4to3 := bn256.PrivatePolyEval(privCoefs4, 3)
	share4to4 := bn256.PrivatePolyEval(privCoefs4, 4)

	groupShares := make([][]byte, 4)
	for k := 0; k < len(groupShares); k++ {
		groupShares[k] = make([]byte, len(mpk.Marshal()))
	}

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := bn256.GenerateGroupSecretKeyPortion(listOfSS1)
	gpk1 := new(bn256.G2).ScalarBaseMult(gsk1)
	groupShares[0] = gpk1.Marshal()
	s1 := new(crypto.BNGroupSigner)
	err := s1.SetPrivk(gsk1.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig1, err := s1.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	listOfSS2 := []*big.Int{share1to2, share2to2, share3to2, share4to2}
	gsk2 := bn256.GenerateGroupSecretKeyPortion(listOfSS2)
	gpk2 := new(bn256.G2).ScalarBaseMult(gsk2)
	groupShares[1] = gpk2.Marshal()
	s2 := new(crypto.BNGroupSigner)
	err = s2.SetPrivk(gsk2.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig2, err := s2.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	listOfSS3 := []*big.Int{share1to3, share2to3, share3to3, share4to3}
	gsk3 := bn256.GenerateGroupSecretKeyPortion(listOfSS3)
	gpk3 := new(bn256.G2).ScalarBaseMult(gsk3)
	groupShares[2] = gpk3.Marshal()
	s3 := new(crypto.BNGroupSigner)
	err = s3.SetPrivk(gsk3.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig3, err := s3.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	listOfSS4 := []*big.Int{share1to4, share2to4, share3to4, share4to4}
	gsk4 := bn256.GenerateGroupSecretKeyPortion(listOfSS4)
	gpk4 := new(bn256.G2).ScalarBaseMult(gsk4)
	groupShares[3] = gpk4.Marshal()
	s4 := new(crypto.BNGroupSigner)
	err = s4.SetPrivk(gsk4.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig4, err := s4.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	sigs := make([][]byte, 4)
	for k := 0; k < len(sigs); k++ {
		sigs[k] = make([]byte, 192)
	}
	sigs[0] = sig1
	sigs[1] = sig2
	sigs[2] = sig3
	sigs[3] = sig4

	err = s.SetGroupPubk(mpk.Marshal())
	if err != nil {
		t.Fatal(err)
	}

	// Make bad sigs array
	sigsBad := make([][]byte, 2)
	for k := 0; k < len(sigsBad); k++ {
		sigsBad[k] = make([]byte, 192)
	}
	sigsBad[0] = sig1
	sigsBad[1] = sig2
	_, err = s.Aggregate(sigsBad, groupShares)
	if err == nil {
		t.Fatal("Should have raised an error for too few signatures!")
	}

	// Finally submit signature
	grpsig, err := s.Aggregate(sigs, groupShares)
	if err != nil {
		t.Fatal(err)
	}

	bnVal := &crypto.BNGroupValidator{}
	groupk, err := bnVal.PubkeyFromSig(grpsig)
	if err != nil {
		t.Fatal(err)
	}

	bnSigners := []*crypto.BNGroupSigner{}
	bnSigners = append(bnSigners, s1)
	bnSigners = append(bnSigners, s2)
	bnSigners = append(bnSigners, s3)
	bnSigners = append(bnSigners, s4)

	secpSigners := []*crypto.Secp256k1Signer{}
	secpPubks := [][]byte{}
	for _, share := range groupShares {
		signer, pubk := makeSecpSigner(share)
		secpPubks = append(secpPubks, pubk)
		secpSigners = append(secpSigners, signer)
	}

	for _, signer := range bnSigners {
		err := signer.SetGroupPubk(groupk)
		if err != nil {
			t.Fatal(err)
		}
	}

	return groupk, bnSigners, groupShares, secpSigners, secpPubks
}

func makeSecpSigner(seed []byte) (*crypto.Secp256k1Signer, []byte) {
	secpSigner := &crypto.Secp256k1Signer{}
	err := secpSigner.SetPrivk(crypto.Hasher(seed))
	if err != nil {
		panic(err)
	}
	secpKey, _ := secpSigner.Pubkey()
	return secpSigner, secpKey
}

func buildRound(t *testing.T, bnSigners []*crypto.BNGroupSigner, groupSharesOrig [][]byte, secpSigners []*crypto.Secp256k1Signer, height uint32, round uint32, prevBlockOrig []byte) (*objs.BlockHeader, []*objs.Proposal, objs.PreVoteList, []*objs.PreVoteNil, objs.PreCommitList, []*objs.PreCommitNil, objs.NextRoundList, objs.NextHeightList, *objs.BlockHeader) {
	groupShares := make([][]byte, len(groupSharesOrig))
	copy(groupShares, groupSharesOrig)
	prevBlock := utils.CopySlice(prevBlockOrig)
	bnVal := &crypto.BNGroupValidator{}
	secpVal := &crypto.Secp256k1Validator{}
	pl := []*objs.Proposal{}
	pvl := objs.PreVoteList{}
	pvnl := []*objs.PreVoteNil{}
	pcl := objs.PreCommitList{}
	pcnl := []*objs.PreCommitNil{}
	nrl := objs.NextRoundList{}
	nhl := objs.NextHeightList{}
	var bh *objs.BlockHeader
	for i := 0; i < len(secpSigners); i++ {
		rcc := &objs.RClaims{
			Height:    height,
			Round:     round,
			PrevBlock: prevBlock,
			ChainID:   1,
		}
		var rc *objs.RCert
		if round > 1 {
			sigs := [][]byte{}
			rcBytes, err := rcc.MarshalBinary()
			if err != nil {
				t.Fatal(err)
			}
			for _, signer := range bnSigners {
				sig, err := signer.Sign(rcBytes)
				if err != nil {
					t.Fatal(err)
				}
				sigs = append(sigs, sig)
			}
			sigGroup, err := bnSigners[i].Aggregate(sigs, groupShares)
			if err != nil {
				t.Fatal(err)
			}
			rcert := &objs.RCert{
				SigGroup: sigGroup,
				RClaims:  rcc,
			}
			rc = rcert
		} else {
			sigs := [][]byte{}
			rcBytes := prevBlock
			for _, signer := range bnSigners {
				sig, err := signer.Sign(rcBytes)
				if err != nil {
					t.Fatal(err)
				}
				sigs = append(sigs, sig)
			}
			sigGroup, err := bnSigners[i].Aggregate(sigs, groupShares)
			if err != nil {
				t.Fatal(err)
			}
			rcert := &objs.RCert{
				SigGroup: sigGroup,
				RClaims:  rcc,
			}
			rc = rcert
		}
		err := rc.ValidateSignature(bnVal)
		if err != nil {
			t.Fatal(err)
		}
		bc := &objs.BClaims{
			ChainID:    1,
			Height:     height,
			PrevBlock:  prevBlock,
			StateRoot:  prevBlock,
			HeaderRoot: crypto.Hasher([]byte{}),
			TxRoot:     crypto.Hasher([]byte{}),
		}
		p := &objs.Proposal{
			TxHshLst: [][]byte{},
			PClaims: &objs.PClaims{
				BClaims: bc,
				RCert:   rc,
			},
		}
		err = p.Sign(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pBytes, err := p.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		p = &objs.Proposal{}
		err = p.UnmarshalBinary(pBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = p.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pl = append(pl, p)
		pv, err := p.PreVote(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pvBytes, err := pv.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		pv = &objs.PreVote{}
		err = pv.UnmarshalBinary(pvBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = pv.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pvl = append(pvl, pv)
		pvn, err := rc.PreVoteNil(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pvnBytes, err := pvn.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		pvn = &objs.PreVoteNil{}
		err = pvn.UnmarshalBinary(pvnBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = pvn.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pvnl = append(pvnl, pvn)
		pcn, err := rc.PreCommitNil(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pcnBytes, err := pcn.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		pcn = &objs.PreCommitNil{}
		err = pcn.UnmarshalBinary(pcnBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = pcn.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pcnl = append(pcnl, pcn)
		nr, err := rc.NextRound(secpSigners[i], bnSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		nrBytes, err := nr.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		nr = &objs.NextRound{}
		err = nr.UnmarshalBinary(nrBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = nr.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		nrl = append(nrl, nr)
		bheader := &objs.BlockHeader{
			TxHshLst: [][]byte{},
			BClaims:  bc,
		}
		bhsh, err := bheader.BlockHash()
		if err != nil {
			t.Fatal(err)
		}
		sigs := [][]byte{}
		for _, signer := range bnSigners {
			sig, err := signer.Sign(bhsh)
			if err != nil {
				t.Fatal(err)
			}
			sigs = append(sigs, sig)
		}
		sigGroup, err := bnSigners[i].Aggregate(sigs, groupShares)
		if err != nil {
			t.Fatal(err)
		}
		bheader.SigGroup = sigGroup
		bh = bheader
		err = bh.ValidateSignatures(bnVal)
		if err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < len(secpSigners); i++ {
		pcc, err := pvl.MakePreCommit(secpSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		pccBytes, err := pcc.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		pcc = &objs.PreCommit{}
		err = pcc.UnmarshalBinary(pccBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = pcc.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		pcl = append(pcl, pcc)
	}
	for i := 0; i < len(secpSigners); i++ {
		nh, err := pcl.MakeNextHeight(secpSigners[i], bnSigners[i])
		if err != nil {
			t.Fatal(err)
		}
		nhBytes, err := nh.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		nh = &objs.NextHeight{}
		err = nh.UnmarshalBinary(nhBytes)
		if err != nil {
			t.Fatal(err)
		}
		err = nh.ValidateSignatures(secpVal, bnVal)
		if err != nil {
			t.Fatal(err)
		}
		nhl = append(nhl, nh)
	}
	newBh, _, err := nhl.MakeBlockHeader(bnSigners[0], groupShares)
	if err != nil {
		t.Fatal(err)
	}
	return bh, pl, pvl, pvnl, pcl, pcnl, nrl, nhl, newBh
}
