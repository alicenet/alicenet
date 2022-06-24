package gossip

/*
import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/dgraph-io/badger/v2"
)

func makeOnes(s []byte) []byte {
	for i := 0; i < len(s); i++ {
		s[i] = 255
	}
	return s
}

func mockGossipFunc(fn func(interfaces.PeerLease) error) {
}

func TestClientHeight2(t *testing.T) {
	// Open the DB.
	bcastCount := 0
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gfn := func(fn func(interfaces.PeerLease) error) {
		bcastCount++
		mockGossipFunc(fn)
	}
	client := &Client{
		DB:              db,
		GossipFunc:      gfn,
		GetTxsForGossip: func(txn *badger.Txn, height uint32) ([]interfaces.Transaction, error) { return nil, nil },
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
	tx := []byte("transaction")
	rcc := &objs.RClaims{
		Height:    2,
		Round:     1,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rccNR := &objs.RClaims{
		Height:    2,
		Round:     2,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rc := &objs.RCert{
		SigGroup: make([]byte, 192),
		RClaims:  rcc,
	}
	bc := &objs.BClaims{
		ChainID:    1,
		Height:     2,
		PrevBlock:  make([]byte, 32),
		StateRoot:  make([]byte, 32),
		HeaderRoot: make([]byte, 32),
		TxRoot:     make([]byte, 32),
	}
	p := &objs.Proposal{
		TxHshLst:  [][]byte{},
		Signature: makeOnes(make([]byte, 65)),
		PClaims: &objs.PClaims{
			BClaims: bc,
			RCert:   rc,
		},
	}
	pv := &objs.PreVote{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
	}
	pvn := &objs.PreVoteNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	pc := &objs.PreCommit{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
		PreVotes:  [][]byte{makeOnes(make([]byte, 65))},
	}
	pcn := &objs.PreCommitNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	nr := &objs.NextRound{
		NRClaims: &objs.NRClaims{
			RCert:    rc,
			RClaims:  rccNR,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature: makeOnes(make([]byte, 65)),
	}
	nh := &objs.NextHeight{
		NHClaims: &objs.NHClaims{
			Proposal: p,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature:  makeOnes(make([]byte, 65)),
		PreCommits: [][]byte{makeOnes(make([]byte, 65))},
	}
	bh := &objs.BlockHeader{
		TxHshLst: [][]byte{},
		BClaims:  bc,
		SigGroup: makeOnes(make([]byte, 192)),
	}
	ownState := &objs.OwnState{
		VAddr:             makeOnes(make([]byte, 20)),
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
	val := &objs.Validator{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupShare: makeOnes(make([]byte, 128)),
	}
	vs := &objs.ValidatorSet{
		Validators: []*objs.Validator{val},
		GroupKey:   makeOnes(make([]byte, 128)),
		NotBefore:  1,
	}
	rs := &objs.RoundState{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupKey:   makeOnes(make([]byte, 128)),
		GroupShare: makeOnes(make([]byte, 128)),
		GroupIdx:   0,
		RCert:      rc,
	}
	_ = p
	_ = pv
	_ = pvn
	_ = pc
	_ = pcn
	_ = nr
	_ = nh
	_ = bc
	err = db.Update(func(txn *badger.Txn) error {
		err = client.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetOwnState(txn, ownState)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastTransaction(txn, tx)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastProposal(txn, p)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastPreVote(txn, pv)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastPreVoteNil(txn, pvn)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastPreCommit(txn, pc)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastPreCommitNil(txn, pcn)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastNextRound(txn, nr)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastNextHeight(txn, nh)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastBlockHeader(txn, bh)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if bcastCount != 9 {
		t.Fatalf("#1 missing a gossip: got %d\n", bcastCount)
	}
	err = client.ReGossip()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if bcastCount != 16 {
		t.Fatalf("#2: missing a gossip: got %d\n", bcastCount)
	}
	client.Exit()
	<-client.Done()
}

func TestClientHeight1Proposal(t *testing.T) {
	// Open the DB.
	bcastCount := 0
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gfn := func(fn func(interfaces.PeerLease) error) {
		bcastCount++
		mockGossipFunc(fn)
	}
	client := &Client{
		DB:              db,
		GossipFunc:      gfn,
		GetTxsForGossip: func(txn *badger.Txn, height uint32) ([]interfaces.Transaction, error) { return nil, nil },
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
	tx := []byte("transaction")
	rcc := &objs.RClaims{
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rccNR := &objs.RClaims{
		Height:    1,
		Round:     2,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rc := &objs.RCert{
		SigGroup: make([]byte, 192),
		RClaims:  rcc,
	}
	bc := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		PrevBlock:  make([]byte, 32),
		StateRoot:  make([]byte, 32),
		HeaderRoot: make([]byte, 32),
		TxRoot:     make([]byte, 32),
	}
	p := &objs.Proposal{
		TxHshLst:  [][]byte{},
		Signature: makeOnes(make([]byte, 65)),
		PClaims: &objs.PClaims{
			BClaims: bc,
			RCert:   rc,
		},
	}
	pv := &objs.PreVote{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
	}
	pvn := &objs.PreVoteNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	pc := &objs.PreCommit{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
		PreVotes:  [][]byte{makeOnes(make([]byte, 65))},
	}
	pcn := &objs.PreCommitNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	nr := &objs.NextRound{
		NRClaims: &objs.NRClaims{
			RCert:    rc,
			RClaims:  rccNR,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature: makeOnes(make([]byte, 65)),
	}
	nh := &objs.NextHeight{
		NHClaims: &objs.NHClaims{
			Proposal: p,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature:  makeOnes(make([]byte, 65)),
		PreCommits: [][]byte{makeOnes(make([]byte, 65))},
	}
	bh := &objs.BlockHeader{
		TxHshLst: [][]byte{},
		BClaims:  bc,
		SigGroup: makeOnes(make([]byte, 192)),
	}
	ownState := &objs.OwnState{
		VAddr:             makeOnes(make([]byte, 20)),
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
	val := &objs.Validator{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupShare: makeOnes(make([]byte, 128)),
	}
	vs := &objs.ValidatorSet{
		Validators: []*objs.Validator{val},
		GroupKey:   makeOnes(make([]byte, 128)),
		NotBefore:  1,
	}
	rs := &objs.RoundState{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupKey:   makeOnes(make([]byte, 128)),
		GroupShare: makeOnes(make([]byte, 128)),
		GroupIdx:   0,
		RCert:      rc,
	}
	_ = p
	_ = pv
	_ = pvn
	_ = pc
	_ = pcn
	_ = nr
	_ = nh
	_ = bc
	err = db.Update(func(txn *badger.Txn) error {
		err = client.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetOwnState(txn, ownState)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastTransaction(txn, tx)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastProposal(txn, p)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if bcastCount != 2 {
		t.Fatalf("#1 missing a gossip: got %d\n", bcastCount)
	}
	err = client.ReGossip()
	if err == nil {
		t.Fatal("Should have raised error!")
	}
	client.Exit()
	<-client.Done()
}

func TestClientHeight1PreVote(t *testing.T) {
	// Open the DB.
	bcastCount := 0
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gfn := func(fn func(interfaces.PeerLease) error) {
		bcastCount++
		mockGossipFunc(fn)
	}
	client := &Client{
		DB:              db,
		GossipFunc:      gfn,
		GetTxsForGossip: func(txn *badger.Txn, height uint32) ([]interfaces.Transaction, error) { return nil, nil },
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
	tx := []byte("transaction")
	rcc := &objs.RClaims{
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rccNR := &objs.RClaims{
		Height:    1,
		Round:     2,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rc := &objs.RCert{
		SigGroup: make([]byte, 192),
		RClaims:  rcc,
	}
	bc := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		PrevBlock:  make([]byte, 32),
		StateRoot:  make([]byte, 32),
		HeaderRoot: make([]byte, 32),
		TxRoot:     make([]byte, 32),
	}
	p := &objs.Proposal{
		TxHshLst:  [][]byte{},
		Signature: makeOnes(make([]byte, 65)),
		PClaims: &objs.PClaims{
			BClaims: bc,
			RCert:   rc,
		},
	}
	pv := &objs.PreVote{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
	}
	pvn := &objs.PreVoteNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	pc := &objs.PreCommit{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
		PreVotes:  [][]byte{makeOnes(make([]byte, 65))},
	}
	pcn := &objs.PreCommitNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	nr := &objs.NextRound{
		NRClaims: &objs.NRClaims{
			RCert:    rc,
			RClaims:  rccNR,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature: makeOnes(make([]byte, 65)),
	}
	nh := &objs.NextHeight{
		NHClaims: &objs.NHClaims{
			Proposal: p,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature:  makeOnes(make([]byte, 65)),
		PreCommits: [][]byte{makeOnes(make([]byte, 65))},
	}
	bh := &objs.BlockHeader{
		TxHshLst: [][]byte{},
		BClaims:  bc,
		SigGroup: makeOnes(make([]byte, 192)),
	}
	ownState := &objs.OwnState{
		VAddr:             makeOnes(make([]byte, 20)),
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
	val := &objs.Validator{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupShare: makeOnes(make([]byte, 128)),
	}
	vs := &objs.ValidatorSet{
		Validators: []*objs.Validator{val},
		GroupKey:   makeOnes(make([]byte, 128)),
		NotBefore:  1,
	}
	rs := &objs.RoundState{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupKey:   makeOnes(make([]byte, 128)),
		GroupShare: makeOnes(make([]byte, 128)),
		GroupIdx:   0,
		RCert:      rc,
	}
	_ = p
	_ = pv
	_ = pvn
	_ = pc
	_ = pcn
	_ = nr
	_ = nh
	_ = bc
	err = db.Update(func(txn *badger.Txn) error {
		err = client.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetOwnState(txn, ownState)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastTransaction(txn, tx)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastPreVote(txn, pv)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if bcastCount != 2 {
		t.Fatalf("#1 missing a gossip: got %d\n", bcastCount)
	}
	err = client.ReGossip()
	if err == nil {
		t.Fatal("Should have raised error!")
	}
	client.Exit()
	<-client.Done()
}

func TestClientHeight1PreVoteNil(t *testing.T) {
	// Open the DB.
	bcastCount := 0
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gfn := func(fn func(interfaces.PeerLease) error) {
		bcastCount++
		mockGossipFunc(fn)
	}
	client := &Client{
		DB:              db,
		GossipFunc:      gfn,
		GetTxsForGossip: func(txn *badger.Txn, height uint32) ([]interfaces.Transaction, error) { return nil, nil },
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
	tx := []byte("transaction")
	rcc := &objs.RClaims{
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rccNR := &objs.RClaims{
		Height:    1,
		Round:     2,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rc := &objs.RCert{
		SigGroup: make([]byte, 192),
		RClaims:  rcc,
	}
	bc := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		PrevBlock:  make([]byte, 32),
		StateRoot:  make([]byte, 32),
		HeaderRoot: make([]byte, 32),
		TxRoot:     make([]byte, 32),
	}
	p := &objs.Proposal{
		TxHshLst:  [][]byte{},
		Signature: makeOnes(make([]byte, 65)),
		PClaims: &objs.PClaims{
			BClaims: bc,
			RCert:   rc,
		},
	}
	pv := &objs.PreVote{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
	}
	pvn := &objs.PreVoteNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	pc := &objs.PreCommit{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
		PreVotes:  [][]byte{makeOnes(make([]byte, 65))},
	}
	pcn := &objs.PreCommitNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	nr := &objs.NextRound{
		NRClaims: &objs.NRClaims{
			RCert:    rc,
			RClaims:  rccNR,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature: makeOnes(make([]byte, 65)),
	}
	nh := &objs.NextHeight{
		NHClaims: &objs.NHClaims{
			Proposal: p,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature:  makeOnes(make([]byte, 65)),
		PreCommits: [][]byte{makeOnes(make([]byte, 65))},
	}
	bh := &objs.BlockHeader{
		TxHshLst: [][]byte{},
		BClaims:  bc,
		SigGroup: makeOnes(make([]byte, 192)),
	}
	ownState := &objs.OwnState{
		VAddr:             makeOnes(make([]byte, 20)),
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
	val := &objs.Validator{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupShare: makeOnes(make([]byte, 128)),
	}
	vs := &objs.ValidatorSet{
		Validators: []*objs.Validator{val},
		GroupKey:   makeOnes(make([]byte, 128)),
		NotBefore:  1,
	}
	rs := &objs.RoundState{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupKey:   makeOnes(make([]byte, 128)),
		GroupShare: makeOnes(make([]byte, 128)),
		GroupIdx:   0,
		RCert:      rc,
	}
	_ = p
	_ = pv
	_ = pvn
	_ = pc
	_ = pcn
	_ = nr
	_ = nh
	_ = bc
	err = db.Update(func(txn *badger.Txn) error {
		err = client.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetOwnState(txn, ownState)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastTransaction(txn, tx)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastPreVoteNil(txn, pvn)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if bcastCount != 2 {
		t.Fatalf("#1 missing a gossip: got %d\n", bcastCount)
	}
	err = client.ReGossip()
	if err == nil {
		t.Fatal("Should have raised error!")
	}
	client.Exit()
	<-client.Done()
}

func TestClientHeight1PreCommit(t *testing.T) {
	// Open the DB.
	bcastCount := 0
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gfn := func(fn func(interfaces.PeerLease) error) {
		bcastCount++
		mockGossipFunc(fn)
	}
	client := &Client{
		DB:              db,
		GossipFunc:      gfn,
		GetTxsForGossip: func(txn *badger.Txn, height uint32) ([]interfaces.Transaction, error) { return nil, nil },
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
	tx := []byte("transaction")
	rcc := &objs.RClaims{
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rccNR := &objs.RClaims{
		Height:    1,
		Round:     2,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rc := &objs.RCert{
		SigGroup: make([]byte, 192),
		RClaims:  rcc,
	}
	bc := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		PrevBlock:  make([]byte, 32),
		StateRoot:  make([]byte, 32),
		HeaderRoot: make([]byte, 32),
		TxRoot:     make([]byte, 32),
	}
	p := &objs.Proposal{
		TxHshLst:  [][]byte{},
		Signature: makeOnes(make([]byte, 65)),
		PClaims: &objs.PClaims{
			BClaims: bc,
			RCert:   rc,
		},
	}
	pv := &objs.PreVote{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
	}
	pvn := &objs.PreVoteNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	pc := &objs.PreCommit{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
		PreVotes:  [][]byte{makeOnes(make([]byte, 65))},
	}
	pcn := &objs.PreCommitNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	nr := &objs.NextRound{
		NRClaims: &objs.NRClaims{
			RCert:    rc,
			RClaims:  rccNR,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature: makeOnes(make([]byte, 65)),
	}
	nh := &objs.NextHeight{
		NHClaims: &objs.NHClaims{
			Proposal: p,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature:  makeOnes(make([]byte, 65)),
		PreCommits: [][]byte{makeOnes(make([]byte, 65))},
	}
	bh := &objs.BlockHeader{
		TxHshLst: [][]byte{},
		BClaims:  bc,
		SigGroup: makeOnes(make([]byte, 192)),
	}
	ownState := &objs.OwnState{
		VAddr:             makeOnes(make([]byte, 20)),
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
	val := &objs.Validator{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupShare: makeOnes(make([]byte, 128)),
	}
	vs := &objs.ValidatorSet{
		Validators: []*objs.Validator{val},
		GroupKey:   makeOnes(make([]byte, 128)),
		NotBefore:  1,
	}
	rs := &objs.RoundState{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupKey:   makeOnes(make([]byte, 128)),
		GroupShare: makeOnes(make([]byte, 128)),
		GroupIdx:   0,
		RCert:      rc,
	}
	_ = p
	_ = pv
	_ = pvn
	_ = pc
	_ = pcn
	_ = nr
	_ = nh
	_ = bc
	err = db.Update(func(txn *badger.Txn) error {
		err = client.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetOwnState(txn, ownState)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastTransaction(txn, tx)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastPreCommit(txn, pc)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if bcastCount != 2 {
		t.Fatalf("#1 missing a gossip: got %d\n", bcastCount)
	}
	err = client.ReGossip()
	if err == nil {
		t.Fatal("Should have raised error!")
	}
	client.Exit()
	<-client.Done()
}

func TestClientHeight1PreCommitNil(t *testing.T) {
	// Open the DB.
	bcastCount := 0
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gfn := func(fn func(interfaces.PeerLease) error) {
		bcastCount++
		mockGossipFunc(fn)
	}
	client := &Client{
		DB:              db,
		GossipFunc:      gfn,
		GetTxsForGossip: func(txn *badger.Txn, height uint32) ([]interfaces.Transaction, error) { return nil, nil },
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
	tx := []byte("transaction")
	rcc := &objs.RClaims{
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rccNR := &objs.RClaims{
		Height:    1,
		Round:     2,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rc := &objs.RCert{
		SigGroup: make([]byte, 192),
		RClaims:  rcc,
	}
	bc := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		PrevBlock:  make([]byte, 32),
		StateRoot:  make([]byte, 32),
		HeaderRoot: make([]byte, 32),
		TxRoot:     make([]byte, 32),
	}
	p := &objs.Proposal{
		TxHshLst:  [][]byte{},
		Signature: makeOnes(make([]byte, 65)),
		PClaims: &objs.PClaims{
			BClaims: bc,
			RCert:   rc,
		},
	}
	pv := &objs.PreVote{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
	}
	pvn := &objs.PreVoteNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	pc := &objs.PreCommit{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
		PreVotes:  [][]byte{makeOnes(make([]byte, 65))},
	}
	pcn := &objs.PreCommitNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	nr := &objs.NextRound{
		NRClaims: &objs.NRClaims{
			RCert:    rc,
			RClaims:  rccNR,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature: makeOnes(make([]byte, 65)),
	}
	nh := &objs.NextHeight{
		NHClaims: &objs.NHClaims{
			Proposal: p,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature:  makeOnes(make([]byte, 65)),
		PreCommits: [][]byte{makeOnes(make([]byte, 65))},
	}
	bh := &objs.BlockHeader{
		TxHshLst: [][]byte{},
		BClaims:  bc,
		SigGroup: makeOnes(make([]byte, 192)),
	}
	ownState := &objs.OwnState{
		VAddr:             makeOnes(make([]byte, 20)),
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
	val := &objs.Validator{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupShare: makeOnes(make([]byte, 128)),
	}
	vs := &objs.ValidatorSet{
		Validators: []*objs.Validator{val},
		GroupKey:   makeOnes(make([]byte, 128)),
		NotBefore:  1,
	}
	rs := &objs.RoundState{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupKey:   makeOnes(make([]byte, 128)),
		GroupShare: makeOnes(make([]byte, 128)),
		GroupIdx:   0,
		RCert:      rc,
	}
	_ = p
	_ = pv
	_ = pvn
	_ = pc
	_ = pcn
	_ = nr
	_ = nh
	_ = bc
	err = db.Update(func(txn *badger.Txn) error {
		err = client.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetOwnState(txn, ownState)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastTransaction(txn, tx)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastPreCommitNil(txn, pcn)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if bcastCount != 2 {
		t.Fatalf("#1 missing a gossip: got %d\n", bcastCount)
	}
	err = client.ReGossip()
	if err == nil {
		t.Fatal("Should have raised error!")
	}
	client.Exit()
	<-client.Done()
}

func TestClientHeight1NextRound(t *testing.T) {
	// Open the DB.
	bcastCount := 0
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gfn := func(fn func(interfaces.PeerLease) error) {
		bcastCount++
		mockGossipFunc(fn)
	}
	client := &Client{
		DB:              db,
		GossipFunc:      gfn,
		GetTxsForGossip: func(txn *badger.Txn, height uint32) ([]interfaces.Transaction, error) { return nil, nil },
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
	tx := []byte("transaction")
	rcc := &objs.RClaims{
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rccNR := &objs.RClaims{
		Height:    1,
		Round:     2,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rc := &objs.RCert{
		SigGroup: make([]byte, 192),
		RClaims:  rcc,
	}
	bc := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		PrevBlock:  make([]byte, 32),
		StateRoot:  make([]byte, 32),
		HeaderRoot: make([]byte, 32),
		TxRoot:     make([]byte, 32),
	}
	p := &objs.Proposal{
		TxHshLst:  [][]byte{},
		Signature: makeOnes(make([]byte, 65)),
		PClaims: &objs.PClaims{
			BClaims: bc,
			RCert:   rc,
		},
	}
	pv := &objs.PreVote{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
	}
	pvn := &objs.PreVoteNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	pc := &objs.PreCommit{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
		PreVotes:  [][]byte{makeOnes(make([]byte, 65))},
	}
	pcn := &objs.PreCommitNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	nr := &objs.NextRound{
		NRClaims: &objs.NRClaims{
			RCert:    rc,
			RClaims:  rccNR,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature: makeOnes(make([]byte, 65)),
	}
	nh := &objs.NextHeight{
		NHClaims: &objs.NHClaims{
			Proposal: p,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature:  makeOnes(make([]byte, 65)),
		PreCommits: [][]byte{makeOnes(make([]byte, 65))},
	}
	bh := &objs.BlockHeader{
		TxHshLst: [][]byte{},
		BClaims:  bc,
		SigGroup: makeOnes(make([]byte, 192)),
	}
	ownState := &objs.OwnState{
		VAddr:             makeOnes(make([]byte, 20)),
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
	val := &objs.Validator{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupShare: makeOnes(make([]byte, 128)),
	}
	vs := &objs.ValidatorSet{
		Validators: []*objs.Validator{val},
		GroupKey:   makeOnes(make([]byte, 128)),
		NotBefore:  1,
	}
	rs := &objs.RoundState{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupKey:   makeOnes(make([]byte, 128)),
		GroupShare: makeOnes(make([]byte, 128)),
		GroupIdx:   0,
		RCert:      rc,
	}
	_ = p
	_ = pv
	_ = pvn
	_ = pc
	_ = pcn
	_ = nr
	_ = nh
	_ = bc
	err = db.Update(func(txn *badger.Txn) error {
		err = client.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetOwnState(txn, ownState)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastTransaction(txn, tx)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastNextRound(txn, nr)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if bcastCount != 2 {
		t.Fatalf("#1 missing a gossip: got %d\n", bcastCount)
	}
	err = client.ReGossip()
	if err == nil {
		t.Fatal("Should have raised error!")
	}
	client.Exit()
	<-client.Done()
}

func TestClientHeight1NextHeight(t *testing.T) {
	// Open the DB.
	bcastCount := 0
	dir, err := ioutil.TempDir("", "badger-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			t.Fatal(err)
		}
	}()
	opts := badger.DefaultOptions(dir)
	db, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	gfn := func(fn func(interfaces.PeerLease) error) {
		bcastCount++
		mockGossipFunc(fn)
	}
	client := &Client{
		DB:              db,
		GossipFunc:      gfn,
		GetTxsForGossip: func(txn *badger.Txn, height uint32) ([]interfaces.Transaction, error) { return nil, nil },
	}
	err = client.Init()
	if err != nil {
		t.Fatal(err)
	}
	err = client.Start()
	if err != nil {
		t.Fatal(err)
	}
	tx := []byte("transaction")
	rcc := &objs.RClaims{
		Height:    1,
		Round:     1,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rccNR := &objs.RClaims{
		Height:    1,
		Round:     2,
		PrevBlock: make([]byte, 32),
		ChainID:   1,
	}
	rc := &objs.RCert{
		SigGroup: make([]byte, 192),
		RClaims:  rcc,
	}
	bc := &objs.BClaims{
		ChainID:    1,
		Height:     1,
		PrevBlock:  make([]byte, 32),
		StateRoot:  make([]byte, 32),
		HeaderRoot: make([]byte, 32),
		TxRoot:     make([]byte, 32),
	}
	p := &objs.Proposal{
		TxHshLst:  [][]byte{},
		Signature: makeOnes(make([]byte, 65)),
		PClaims: &objs.PClaims{
			BClaims: bc,
			RCert:   rc,
		},
	}
	pv := &objs.PreVote{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
	}
	pvn := &objs.PreVoteNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	pc := &objs.PreCommit{
		Proposal:  p,
		Signature: makeOnes(make([]byte, 65)),
		PreVotes:  [][]byte{makeOnes(make([]byte, 65))},
	}
	pcn := &objs.PreCommitNil{
		RCert:     rc,
		Signature: makeOnes(make([]byte, 65)),
	}
	nr := &objs.NextRound{
		NRClaims: &objs.NRClaims{
			RCert:    rc,
			RClaims:  rccNR,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature: makeOnes(make([]byte, 65)),
	}
	nh := &objs.NextHeight{
		NHClaims: &objs.NHClaims{
			Proposal: p,
			SigShare: makeOnes(make([]byte, 192)),
		},
		Signature:  makeOnes(make([]byte, 65)),
		PreCommits: [][]byte{makeOnes(make([]byte, 65))},
	}
	bh := &objs.BlockHeader{
		TxHshLst: [][]byte{},
		BClaims:  bc,
		SigGroup: makeOnes(make([]byte, 192)),
	}
	os := &objs.OwnState{
		VAddr:             makeOnes(make([]byte, 20)),
		SyncToBH:          bh,
		MaxBHSeen:         bh,
		CanonicalSnapShot: bh,
		PendingSnapShot:   bh,
	}
	val := &objs.Validator{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupShare: makeOnes(make([]byte, 128)),
	}
	vs := &objs.ValidatorSet{
		Validators: []*objs.Validator{val},
		GroupKey:   makeOnes(make([]byte, 128)),
		NotBefore:  1,
	}
	rs := &objs.RoundState{
		VAddr:      makeOnes(make([]byte, 20)),
		GroupKey:   makeOnes(make([]byte, 128)),
		GroupShare: makeOnes(make([]byte, 128)),
		GroupIdx:   0,
		RCert:      rc,
	}
	_ = p
	_ = pv
	_ = pvn
	_ = pc
	_ = pcn
	_ = nr
	_ = nh
	_ = bc
	err = db.Update(func(txn *badger.Txn) error {
		err = client.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastTransaction(txn, tx)
		if err != nil {
			t.Fatal(err)
		}
		err = client.database.SetBroadcastNextHeight(txn, nh)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	if bcastCount != 2 {
		t.Fatalf("#1 missing a gossip: got %d\n", bcastCount)
	}
	err = client.ReGossip()
	if err == nil {
		t.Fatal("Should have raised error!")
	}
	client.Exit()
	<-client.Done()
}
*/
