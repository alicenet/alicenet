package lstate

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/MadBase/MadNet/application"
	"github.com/MadBase/MadNet/consensus/admin"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/consensus/request"
	hashlib "github.com/MadBase/MadNet/crypto"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/crypto"
)

func getStateHandler(t *testing.T) *Engine {
	stateHandler := &Engine{}
	conDB := &db.Database{}
	dman := &DMan{}
	app := &application.Application{}
	cesigner := &hashlib.Secp256k1Signer{}
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

	rbusClient := &request.Client{}

	if err := stateHandler.Init(conDB, dman, app, cesigner, ah, publicKey, rbusClient); err != nil {
		panic(err)
	}

	return stateHandler
}

func TestCriteria(t *testing.T) {

	stateHandler := getStateHandler(t)

	roundStates := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
		ValidatorSet:       &objs.ValidatorSet{Validators: []*objs.Validator{&objs.Validator{}}},
		OwnValidatingState: &objs.OwnValidatingState{}}

	type testStruct struct {
		h                    []handler
		expectedCaseSelected int
	}

	tests := []testStruct{
		{h: []handler{
			nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: true},
			pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: false},
			pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: false},
			pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: false},
			pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: false},
			ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 0},
		{h: []handler{
			nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: false},
			pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: true},
			pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: false},
			pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: false},
			pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: false},
			ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 1},
		{h: []handler{
			nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: false},
			pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: false},
			pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: true},
			pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: false},
			pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: false},
			ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 2},
		{h: []handler{
			nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: false},
			pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: false},
			pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: false},
			pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: true},
			pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: false},
			ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 3},
		{h: []handler{
			nrCurrentHandler{ce: stateHandler, rs: roundStates, NRCurrent: false},
			pcCurrentHandler{ce: stateHandler, rs: roundStates, PCCurrent: false},
			pcnCurrentHandler{ce: stateHandler, rs: roundStates, PCNCurrent: false},
			pvCurrentHandler{ce: stateHandler, rs: roundStates, PVCurrent: false},
			pvnCurrentHandler{ce: stateHandler, rs: roundStates, PVNCurrent: true},
			ptoExpiredHandler{ce: stateHandler, rs: roundStates},
			validPropHandler{ce: stateHandler, rs: roundStates, PCurrent: false},
		}, expectedCaseSelected: 4},
		// don't have a test for valid prop handler right now because would have to set things up
		// so that the is proposer function returns true. could probably set that up at some point
	}

	for i := 0; i < len(tests); i++ {
		caseSelected := 0
		fn := func(txn *badger.Txn) error {

			for j := 0; j < len(tests[i].h); j++ {
				tests[i].h[j].setTxn(txn)
			}

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

func TestFhFunc(t *testing.T) {
	stateHandler := getStateHandler(t)

	fn := func(txn *badger.Txn) error {
		roundStates := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
			ValidatorSet: &objs.ValidatorSet{}, OwnValidatingState: &objs.OwnValidatingState{}}

		rs := &objs.RoundState{}

		booleanValue, err := stateHandler.fhFunc(txn, roundStates, rs)
		if err != nil {
			fmt.Println("err is", err)
		}

		fmt.Println("boolean value from fhFunc is", booleanValue)

		return nil
	}
	testDb(t, fn)
}

func TestRCertFunc(t *testing.T) {
	stateHandler := getStateHandler(t)

	fn := func(txn *badger.Txn) error {
		roundStates := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
			ValidatorSet: &objs.ValidatorSet{}, OwnValidatingState: &objs.OwnValidatingState{}}

		roundStates.PeerStateMap = make(map[string]*objs.RoundState)

		maxrCert := &objs.RCert{RClaims: &objs.RClaims{ChainID: 0, Height: 0, Round: 0, PrevBlock: []byte{0}},
			SigGroup: []byte{0}, GroupKey: []byte{0}}

		booleanValue, err := stateHandler.rCertFunc(txn, roundStates, maxrCert)
		if err != nil {
			fmt.Println("err is", err)
		}

		fmt.Println("boolean value from rCertFunc is", booleanValue)

		return nil
	}
	testDb(t, fn)
}

func TestNrCurrent(t *testing.T) {
	stateHandler := getStateHandler(t)

	fn := func(txn *badger.Txn) error {
		roundState := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
			ValidatorSet: &objs.ValidatorSet{}, OwnValidatingState: &objs.OwnValidatingState{}}

		booleanValue, err := stateHandler.nrCurrentFunc(txn, roundState)
		if err != nil {
			fmt.Println("err is", err)
		}

		fmt.Println("boolean value from nrCurrentFunc is", booleanValue)

		return nil
	}
	testDb(t, fn)
}

func TestPcCurrent(t *testing.T) {
	stateHandler := getStateHandler(t)

	fn := func(txn *badger.Txn) error {
		// this is how the round state was being generated in the real system
		// roundState, err := stateHandler.sstore.LoadLocalState(txn)
		// if err != nil {
		// 	return err
		// }
		roundState := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
			ValidatorSet: &objs.ValidatorSet{}, OwnValidatingState: &objs.OwnValidatingState{}}

		// not really sure what the value for this should be
		PCTOExpired := false
		booleanValue, err := stateHandler.pcCurrentFunc(txn, roundState, PCTOExpired)
		if err != nil {
			fmt.Println("err is", err)
		}

		fmt.Println("boolean value from pcCurrentFunc is", booleanValue)

		return nil
	}
	testDb(t, fn)
}

func TestPcnCurrent(t *testing.T) {
	stateHandler := getStateHandler(t)

	fn := func(txn *badger.Txn) error {
		roundState := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
			ValidatorSet: &objs.ValidatorSet{}, OwnValidatingState: &objs.OwnValidatingState{}}

		// not really sure what the value for this should be
		PCTOExpired := true
		booleanValue, err := stateHandler.pcnCurrentFunc(txn, roundState, PCTOExpired)
		if err != nil {
			fmt.Println("err is", err)
		}

		fmt.Println("boolean value from pcnCurrentFunc is", booleanValue)

		return nil
	}
	testDb(t, fn)
}

func TestPvCurrent(t *testing.T) {
	stateHandler := getStateHandler(t)

	fn := func(txn *badger.Txn) error {
		roundState := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
			ValidatorSet: &objs.ValidatorSet{}, OwnValidatingState: &objs.OwnValidatingState{}}

		PVTOExpired := true
		booleanValue, err := stateHandler.pvCurrentFunc(txn, roundState, PVTOExpired)
		if err != nil {
			fmt.Println("err is", err)
		}

		fmt.Println("boolean value from pvCurrentFunc is", booleanValue)

		return nil
	}
	testDb(t, fn)
}

func TestPvnCurrent(t *testing.T) {
	stateHandler := getStateHandler(t)

	fn := func(txn *badger.Txn) error {
		roundState := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
			ValidatorSet: &objs.ValidatorSet{}, OwnValidatingState: &objs.OwnValidatingState{}}

		PVTOExpired := true
		booleanValue, err := stateHandler.pvnCurrentFunc(txn, roundState, PVTOExpired)
		if err != nil {
			fmt.Println("err is", err)
		}

		fmt.Println("boolean value from pvCurrentFunc is", booleanValue)

		return nil
	}
	testDb(t, fn)
}

func TestPtoExpired(t *testing.T) {
	stateHandler := getStateHandler(t)

	fn := func(txn *badger.Txn) error {
		roundState := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
			ValidatorSet: &objs.ValidatorSet{}, OwnValidatingState: &objs.OwnValidatingState{}}

		booleanValue, err := stateHandler.ptoExpiredFunc(txn, roundState)
		if err != nil {
			fmt.Println("err is", err)
		}

		fmt.Println("boolean value from pvCurrentFunc is", booleanValue)

		return nil
	}
	testDb(t, fn)
}

func TestValidProp(t *testing.T) {
	stateHandler := getStateHandler(t)

	fn := func(txn *badger.Txn) error {
		roundState := &RoundStates{height: 0, round: 0, OwnState: &objs.OwnState{},
			ValidatorSet: &objs.ValidatorSet{}, OwnValidatingState: &objs.OwnValidatingState{}}

		booleanValue, err := stateHandler.validPropFunc(txn, roundState)
		if err != nil {
			fmt.Println("err is", err)
		}

		fmt.Println("boolean value from pvCurrentFunc is", booleanValue)

		return nil
	}
	testDb(t, fn)
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
