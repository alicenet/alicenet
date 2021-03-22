package lstate

import (
	"crypto/ecdsa"
	"crypto/rand"
	"io/ioutil"
	"os"
	"testing"

	"github.com/MadBase/MadNet/application/deposit"
	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/MadBase/MadNet/consensus/appmock"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/consensus/request"
	hashlib "github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/peering"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/crypto"
)

func getStateHandler(t *testing.T, mdb db.DatabaseIface) *Engine {

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
	bdb, err := badger.Open(opts)
	if err != nil {
		t.Fatal(err)
	}
	defer bdb.Close()

	dHandler := &deposit.Handler{}
	dHandler.Init()

	stateHandler := &Engine{}

	app := appmock.New()

	dman := &DMan{}
	rb := &request.Client{}

	rb.Init(&peering.PeerSubscription{})
	dman.Init(mdb, app, rb)

	ah := &admin.Handlers{}

	c := crypto.S256()

	priv, err := ecdsa.GenerateKey(c, rand.Reader)
	if err != nil {
		t.Errorf("error: %s", err)
		return nil
	}
	if !c.IsOnCurve(priv.PublicKey.X, priv.PublicKey.Y) {
		t.Errorf("public key invalid: %s", err)
	}

	publicKey := crypto.FromECDSAPub(&priv.PublicKey)

	cesigner := &hashlib.Secp256k1Signer{}
	err = cesigner.SetPrivk(priv.X.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	rbusClient := &request.Client{}

	if err := stateHandler.Init(mdb, dman, app, cesigner, ah, publicKey, rbusClient); err != nil {
		panic(err)
	}

	stateHandler.bnSigner = &hashlib.BNGroupSigner{}
	stateHandler.bnSigner.SetPrivk([]byte{173, 233, 94, 109, 13, 42, 99, 22})

	return stateHandler
}

func TestCriteria(t *testing.T) {

	stateHandler := getStateHandler(t, nil)

	roundStates := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
		ValidatorSet:       &objs.ValidatorSet{Validators: []*objs.Validator{}},
		OwnValidatingState: &objs.OwnValidatingState{}}

	type testStruct struct {
		h                    []handler
		expectedCaseSelected int
	}

	tests := []testStruct{
		{h: []handler{
			&nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: true},
			&pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: false},
			&pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: false},
			&pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: false},
			&pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: false},
			&ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			&validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 0},
		{h: []handler{
			&nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: false},
			&pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: true},
			&pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: false},
			&pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: false},
			&pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: false},
			&ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			&validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 1},
		{h: []handler{
			&nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: false},
			&pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: false},
			&pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: true},
			&pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: false},
			&pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: false},
			&ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			&validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 2},
		{h: []handler{
			&nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: false},
			&pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: false},
			&pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: false},
			&pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: true},
			&pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: false},
			&ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			&validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 3},
		{h: []handler{
			&nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: false},
			&pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: false},
			&pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: false},
			&pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: false},
			&pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: true},
			&ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			&validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 4},
		// don't have a test for valid prop handler right now because would have to set things up
		// so that the is proposer function returns true. could probably set that up at some point
	}

	for i := 0; i < len(tests); i++ {
		caseSelected := 0
		fn := func(txn *badger.Txn) error {
			for j := 0; j < len(tests[i].h); j++ {
				if tests[i].h[j].evalCriteria() {
					caseSelected = j
					break
				}
			}
			return nil
		}
		testDb(t, fn)
		if caseSelected != tests[i].expectedCaseSelected {
			t.Fatal("incorrect handler was selected")
		}
	}

}

func testDb(t *testing.T, fn func(txn *badger.Txn) error) {
	// Open the DB.
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
	err = db.Update(fn)
	if err != nil {
		t.Fatal(err)
	}
}
