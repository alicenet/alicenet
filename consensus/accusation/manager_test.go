package accusation

import (
	"context"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/MadBase/MadNet/consensus/db"
	"github.com/MadBase/MadNet/consensus/objs"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/logging"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type managerTestProxy struct {
	logger         *logrus.Logger
	db             *db.Database
	manager        *Manager
	rawConsensusDb *badger.DB
	validatorSet   *objs.ValidatorSet
}

func generateValidatorSet(t *testing.T) *objs.ValidatorSet {
	gpkj1, ok := big.NewInt(0).SetString("14395602319113363333690669395961581081803242358678131578916981232954633806960", 10)
	assert.True(t, ok)
	gpkj2, ok := big.NewInt(0).SetString("300089735810954642595088127891607498572672898349379085034409445552605516765", 10)
	assert.True(t, ok)
	gpkj3, ok := big.NewInt(0).SetString("17169409825226096532229555694191340178889298261881998623204757401596570351688", 10)
	assert.True(t, ok)
	gpkj4, ok := big.NewInt(0).SetString("19780380227412019371988923760536598779715024137904246485146692590642474692882", 10)
	assert.True(t, ok)

	v1 := objects.Validator{
		Account: common.HexToAddress("0x9AC1c9afBAec85278679fF75Ef109217f26b1417"),
		Index:   1,
		SharedKey: [4]*big.Int{
			gpkj1,
			gpkj2,
			gpkj3,
			gpkj4,
		},
	}

	gpkj1, ok = big.NewInt(0).SetString("21154017404198718862920160130737623556546602199694661996869957208062851500379", 10)
	assert.True(t, ok)
	gpkj2, ok = big.NewInt(0).SetString("19389833000731437962153734187923001234830293448701992540723746507685386979412", 10)
	assert.True(t, ok)
	gpkj3, ok = big.NewInt(0).SetString("21289029302611008572663530729853170393569891172031986702208364730022339833735", 10)
	assert.True(t, ok)
	gpkj4, ok = big.NewInt(0).SetString("15926764275937493411567546154328577890519582979565228998979506880914326856186", 10)
	assert.True(t, ok)

	v2 := objects.Validator{
		Account: common.HexToAddress("0x615695C4a4D6a60830e5fca4901FbA099DF26271"),
		Index:   2,
		SharedKey: [4]*big.Int{
			gpkj1,
			gpkj2,
			gpkj3,
			gpkj4,
		},
	}

	gpkj1, ok = big.NewInt(0).SetString("15079629603150363557558188402860791995814736941924946256968815481986722866449", 10)
	assert.True(t, ok)
	gpkj2, ok = big.NewInt(0).SetString("11164680325282976674805760467491699367894125557056167854003650409966070344792", 10)
	assert.True(t, ok)
	gpkj3, ok = big.NewInt(0).SetString("18616624374737795490811424594534628399519274885945803292205658067710235197668", 10)
	assert.True(t, ok)
	gpkj4, ok = big.NewInt(0).SetString("4331613963825409904165282575933135091483251249365224295595121580000486079984", 10)
	assert.True(t, ok)

	v3 := objects.Validator{
		Account: common.HexToAddress("0x63a6627b79813A7A43829490C4cE409254f64177"),
		Index:   3,
		SharedKey: [4]*big.Int{
			gpkj1,
			gpkj2,
			gpkj3,
			gpkj4,
		},
	}

	gpkj1, ok = big.NewInt(0).SetString("10875965504600753744265546216544158224793678652818595873355677460529088515116", 10)
	assert.True(t, ok)
	gpkj2, ok = big.NewInt(0).SetString("7912658035712558991777053184829906144303269569825235765302768068512975453162", 10)
	assert.True(t, ok)
	gpkj3, ok = big.NewInt(0).SetString("11324169944454120842956077363729540506362078469024985744551121054724657909930", 10)
	assert.True(t, ok)
	gpkj4, ok = big.NewInt(0).SetString("11005450895245397587287710270721947847266013997080161834700568409163476112947", 10)
	assert.True(t, ok)

	v4 := objects.Validator{
		Account: common.HexToAddress("0x16564cF3e880d9F5d09909F51b922941EbBbC24d"),
		Index:   4,
		SharedKey: [4]*big.Int{
			gpkj1,
			gpkj2,
			gpkj3,
			gpkj4,
		},
	}

	validators := []objects.Validator{v1, v2, v3, v4}
	ptrGroupKey := [4]*big.Int{
		v1.SharedKey[0],
		v1.SharedKey[1],
		v1.SharedKey[2],
		v1.SharedKey[3],
	}
	groupKey, err := bn256.MarshalG2Big(ptrGroupKey)
	assert.Nil(t, err)
	vs := &objs.ValidatorSet{
		GroupKey:   groupKey,
		Validators: make([]*objs.Validator, len(validators)),
		NotBefore:  0,
	}

	for _, validator := range validators {
		v := &objs.Validator{
			VAddr:      validator.Account.Bytes(),
			GroupShare: groupKey,
		}
		vs.Validators[validator.Index-1] = v
	}

	return vs
}

func setupManagerTests(t *testing.T) (testProxy *managerTestProxy, closeFn func()) {
	logger := logging.GetLogger("Test")
	deferables := make([]func(), 0)

	closeFn = func() {
		// iterate in reverse order because deferables behave like a stack:
		// the last added deferable should be the first executed
		totalDeferables := len(deferables)
		for i := totalDeferables - 1; i >= 0; i-- {
			deferables[i]()
		}
	}

	ctx := context.Background()
	nodeCtx, cf := context.WithCancel(ctx)
	deferables = append(deferables, cf)

	// Initialize consensus db: stores all state the consensus mechanism requires to work
	rawConsensusDb, err := utils.OpenBadger(nodeCtx.Done(), "", true)
	assert.Nil(t, err)
	var closeDB func() = func() {
		err := rawConsensusDb.Close()
		if err != nil {
			panic(fmt.Errorf("error closing rawConsensusDb: %v", err))
		}
	}
	deferables = append(deferables, closeDB)

	db := &db.Database{}
	db.Init(rawConsensusDb)

	vs := generateValidatorSet(t)

	testProxy = &managerTestProxy{
		logger:         logger,
		db:             db,
		rawConsensusDb: rawConsensusDb,
		validatorSet:   vs,
	}

	testProxy.manager = NewManager(db, logger)
	deferables = append(deferables, testProxy.manager.StopWorkers)

	return
}

func TestManagerStartStop(t *testing.T) {
	testProxy, closeFn := setupManagerTests(t)
	defer closeFn()

	testProxy.manager.StartWorkers()

	time.Sleep(5 * time.Second)
}

func TestManagerPollKeyNotFound(t *testing.T) {
	testProxy, closeFn := setupManagerTests(t)
	defer closeFn()

	testProxy.manager.StartWorkers()

	err := testProxy.manager.Poll()
	assert.NotNil(t, err)

	time.Sleep(5 * time.Second)
}
