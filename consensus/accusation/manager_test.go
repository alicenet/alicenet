package accusation

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"
	"time"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/consensus/lstate"
	"github.com/alicenet/alicenet/consensus/objs"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	bn256 "github.com/alicenet/alicenet/crypto/bn256/cloudflare"
	"github.com/alicenet/alicenet/layer1/executor"
	"github.com/alicenet/alicenet/layer1/executor/marshaller"
	"github.com/alicenet/alicenet/layer1/executor/mocks"
	"github.com/alicenet/alicenet/layer1/executor/tasks"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type managerTestProxy struct {
	logger         *logrus.Logger
	db             *db.Database
	manager        *Manager
	rawConsensusDb *badger.DB
	ta             executor.TaskHandler
}

func setupManagerTests(t *testing.T) *managerTestProxy {
	logger := logging.GetLogger("Test")
	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	t.Cleanup(cf)

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb, err := utils.OpenBadger(nodeCtx.Done(), "", true)
	assert.Nil(t, err)
	var closeDB func() = func() {
		err := rawConsensusDb.Close()
		if err != nil {
			t.Errorf("error closing rawConsensusDb: %v", err)
		}
	}
	t.Cleanup(closeDB)

	db := &db.Database{}
	db.Init(rawConsensusDb)

	sstore := &lstate.Store{}
	sstore.Init(db)

	var ta executor.TaskHandler = mocks.NewMockTaskHandler()

	testProxy := &managerTestProxy{
		logger:         logger,
		db:             db,
		rawConsensusDb: rawConsensusDb,
		ta:             ta,
	}

	testProxy.manager = NewManager(db, sstore, ta, logger)
	t.Cleanup(testProxy.manager.StopWorkers)

	return testProxy
}

// TestManagerStartStop tests the accusation system basic start and stop
func TestManagerStartStop(t *testing.T) {
	testProxy := setupManagerTests(t)
	testProxy.manager.StartWorkers()
}

// TestManagerStopWithoutStart tests the accusation system stop without a start
func TestManagerStopWithoutStart(t *testing.T) {
	_ = setupManagerTests(t)
}

// TestManagerPollKeyNotFound tests the accusation system when there are no round states
func TestManagerPollKeyNotFound(t *testing.T) {
	testProxy := setupManagerTests(t)

	testProxy.manager.StartWorkers()

	err := testProxy.manager.Poll()
	assert.NotNil(t, err)
}

// TestManagerPollCache tests the accusation system caching of round states
func TestManagerPollCache(t *testing.T) {
	testProxy := setupManagerTests(t)

	// no need to start workers because we need to assess if RoundStates are properly polled
	//testProxy.manager.StartWorkers()

	err := testProxy.manager.Poll()
	assert.NotNil(t, err)

	// add validatorSet and roundstates to make it work
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	_ = testProxy.manager.database.Update(func(txn *badger.Txn) error {
		err := testProxy.manager.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = testProxy.manager.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = testProxy.manager.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		result, err := testProxy.manager.sstore.LoadLocalState(txn)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		assert.NotNil(t, result)
		return nil
	})

	// polling should work now
	err = testProxy.manager.Poll()
	assert.Nil(t, err)

	// check if roundstates are in the workQueue
	receivedRS := 0
	//done := false
	var lrs *lstate.RoundStates

	for receivedRS < 1 {
		select {
		case lrs = <-testProxy.manager.workQ:
			receivedRS += 1
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	// assert there is a RS to in the workQ
	assert.Equal(t, 1, receivedRS)
	assert.NotNil(t, lrs)

	// process the RS like a worker would
	hadUpdates, err := testProxy.manager.processLRS(lrs)
	assert.Nil(t, err)
	assert.True(t, hadUpdates)

	// now poll again with the same data.
	// this should not add a new RS to the workQ because it's been processed and cached already
	err = testProxy.manager.Poll()
	assert.Nil(t, err)

	// checking roundstates are in the workQueue
	receivedRS = 0

	for receivedRS < 1 {
		select {
		case <-testProxy.manager.workQ:
			receivedRS += 1
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	assert.Equal(t, 1, receivedRS)

	// process the RS like a worker would, but this time there's nothing new to be processed
	hadUpdates, err = testProxy.manager.processLRS(lrs)
	assert.Nil(t, err)
	assert.False(t, hadUpdates)

	// check cache is invalidated by changing the RS object
	_ = testProxy.manager.database.Update(func(txn *badger.Txn) error {
		rs.RCert.RClaims.Round = 2

		err = testProxy.manager.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		// read the current local roundstate
		result, err := testProxy.manager.sstore.LoadLocalState(txn)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		assert.NotNil(t, result)
		lrs = result

		return nil
	})

	// polling again
	err = testProxy.manager.Poll()
	assert.Nil(t, err)

	// checking roundstates are in the workQueue
	receivedRS = 0

	for receivedRS < 1 {
		select {
		case <-testProxy.manager.workQ:
			receivedRS += 1
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}

	assert.Equal(t, 1, receivedRS)

	// process the RS like a worker would, and this time there are updates
	hadUpdates, err = testProxy.manager.processLRS(lrs)
	assert.Nil(t, err)
	assert.True(t, hadUpdates)
}

// mockAccusationTask
type mockAccusationTask struct {
	*tasks.BaseTask
	SomeData string
}

var _ tasks.Task = &mockAccusationTask{}

func (t *mockAccusationTask) Prepare(ctx context.Context) *tasks.TaskErr {
	return nil
}

func (t *mockAccusationTask) Execute(ctx context.Context) (*types.Transaction, *tasks.TaskErr) {
	return nil, nil
}

func (t *mockAccusationTask) ShouldExecute(ctx context.Context) (bool, *tasks.TaskErr) {
	return true, nil
}

// accuseAllRoundStates is a detector function that accuses all round states because it's a test
func accuseAllRoundStates(rs *objs.RoundState, lrs *lstate.RoundStates, db *db.Database) (tasks.Task, bool) {
	acc := &mockAccusationTask{
		BaseTask: tasks.NewBaseTask(0, 0, false, nil),
		SomeData: "accusing all the things",
	}
	acc.Id = hex.EncodeToString(crypto.Hasher([]byte("some id")))

	return acc, true
}

// assert accuseAllRoundStates is of type detector
var _ detector = accuseAllRoundStates

// TestManagerAccusable tests the accusation system when a malicious behaviour is detected and an accusation is formed
func TestManagerAccusable(t *testing.T) {
	testProxy := setupManagerTests(t)

	// attach accuseAllRoundStates to the manager processing pipeline
	testProxy.manager.detectionPipeline = append(testProxy.manager.detectionPipeline, accuseAllRoundStates)

	// start workers
	testProxy.manager.StartWorkers()

	err := testProxy.manager.Poll()
	assert.NotNil(t, err)

	// add validatorSet and roundstates to make it work
	os := createOwnState(t, 1)
	rs := createRoundState(t, os)
	vs := createValidatorsSet(t, os, rs)

	_ = testProxy.manager.database.Update(func(txn *badger.Txn) error {
		err := testProxy.manager.database.SetOwnState(txn, os)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = testProxy.manager.database.SetCurrentRoundState(txn, rs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		err = testProxy.manager.database.SetValidatorSet(txn, vs)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		result, err := testProxy.manager.sstore.LoadLocalState(txn)
		if err != nil {
			t.Fatalf("Shouldn't have raised error: %v", err)
		}

		assert.NotNil(t, result)
		return nil
	})

	// poll round states
	err = testProxy.manager.Poll()
	assert.Nil(t, err)

	// wait for workers to process the accusation
	time.Sleep(5 * time.Second)

	// check if an accusation is inside the accusationQueue
	receivedAcc := 0
	var accusation tasks.Task
	assert.Nil(t, accusation)

	for receivedAcc < len(vs.Validators) { // all validators are being accused here
		select {
		case acc := <-testProxy.manager.accusationQ:
			t.Logf("received acc: %#v", acc)
			accusation = acc
			receivedAcc += 1

			assert.NotNil(t, accusation)
		default:
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// TestManagerPersistCreatedAccusations tests the persistence of created accusations
func TestManagerPersistCreatedAccusations(t *testing.T) {
	testProxy := setupManagerTests(t)

	// register mockAccusationTask in gob
	gob.Register(&mockAccusationTask{})

	// attach accuseAllRoundStates to the manager processing pipeline
	testProxy.manager.detectionPipeline = append(testProxy.manager.detectionPipeline, accuseAllRoundStates)

	// create accusation
	accusation := &mockAccusationTask{BaseTask: tasks.NewBaseTask(0, 0, true, nil)}
	idString := hex.EncodeToString(crypto.Hasher([]byte("some ID")))
	accusation.Id = idString
	assert.Equal(t, 64, len(accusation.Id))

	assert.Empty(t, testProxy.manager.unpersistedCreatedAccusations)

	// append accusation to the manager's persistence list
	testProxy.manager.unpersistedCreatedAccusations = append(testProxy.manager.unpersistedCreatedAccusations, accusation)

	assert.Equal(t, 1, len(testProxy.manager.unpersistedCreatedAccusations))

	testProxy.manager.persistCreatedAccusations()

	assert.Empty(t, testProxy.manager.unpersistedCreatedAccusations)

	err := testProxy.manager.database.View(func(txn *badger.Txn) error {
		var id [32]byte
		idBin, err := hex.DecodeString(accusation.GetId())
		assert.Nil(t, err)
		copy(id[:], idBin)

		accRaw, err := testProxy.manager.database.GetAccusationRaw(txn, id)
		assert.Nil(t, err)
		acc, err := marshaller.GobUnmarshalBinary(accRaw)
		assert.Nil(t, err)

		assert.Equal(t, acc.GetId(), accusation.GetId())

		accs, err := testProxy.manager.database.GetAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))

		return nil
	})
	assert.Nil(t, err)
}

// TestManagerPersistScheduledAccusations tests the persistence of scheduled accusations
func TestManagerPersistScheduledAccusations(t *testing.T) {
	testProxy := setupManagerTests(t)

	// register mockAccusationTask in gob
	gob.Register(&mockAccusationTask{})

	// attach accuseAllRoundStates to the manager processing pipeline
	testProxy.manager.detectionPipeline = append(testProxy.manager.detectionPipeline, accuseAllRoundStates)

	// create a Persisted accusation and store in DB
	accusation := &mockAccusationTask{
		BaseTask: tasks.NewBaseTask(0, 0, true, nil),
		SomeData: "accusing all the things",
	}
	accusation.Id = hex.EncodeToString(crypto.Hasher([]byte("some ID")))
	var id [32]byte
	idBin, err := hex.DecodeString(accusation.GetId())
	assert.Nil(t, err)
	copy(id[:], idBin)
	accusationRaw, err := marshaller.GobMarshalBinary(accusation)
	assert.Nil(t, err)

	err = testProxy.manager.database.Update(func(txn *badger.Txn) error {
		return testProxy.manager.database.SetAccusationRaw(txn, id, accusationRaw)
	})
	assert.Nil(t, err)

	err = testProxy.manager.scheduleAccusations()
	assert.Nil(t, err)

	err = testProxy.manager.database.View(func(txn *badger.Txn) error {
		accRaw, err := testProxy.manager.database.GetAccusationRaw(txn, id)
		assert.Nil(t, err)
		acc, err := marshaller.GobUnmarshalBinary(accRaw)
		assert.Nil(t, err)

		assert.Equal(t, acc.GetId(), accusation.GetId())

		accs, err := testProxy.manager.database.GetAccusations(txn)
		assert.Nil(t, err)
		assert.Equal(t, 1, len(accs))

		return nil
	})
	assert.Nil(t, err)
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
				Height:    2,
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
