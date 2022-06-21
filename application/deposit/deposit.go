package deposit

import (
	"math/big"

	"github.com/alicenet/alicenet/constants/dbprefix"
	"github.com/alicenet/alicenet/errorz"

	"github.com/alicenet/alicenet/application/db"
	"github.com/alicenet/alicenet/application/indexer"
	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/logging"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"
)

// Handler creates a value owner index of all deposits and allows
// these deposits to be returned for use in a transaction or for verification.
type Handler struct {
	valueIndex *indexer.ValueIndex
	IsSpent    func(txn *badger.Txn, utxoID []byte) (bool, error)
	logger     *logrus.Logger
}

// Init initializes the deposit handler
func (dp *Handler) Init() {
	vidx := indexer.NewValueIndex(
		dbprefix.PrefixDepositValueKey,
		dbprefix.PrefixDepositValueRefKey,
	)
	dp.valueIndex = vidx
	dp.logger = logging.GetLogger(constants.LoggerApp)
}

// IsValid determines if the deposits in txvec are valid
func (dp *Handler) IsValid(txn *badger.Txn, txs objs.TxVec) ([]*objs.TXOut, error) {
	utxoIDs, err := txs.ConsumedUTXOIDOnlyDeposits()
	if err != nil {
		utils.DebugTrace(dp.logger, err)
		return nil, err
	}
	found, missing, spent, err := dp.Get(txn, utxoIDs)
	if err != nil {
		utils.DebugTrace(dp.logger, err)
		return nil, err
	}
	if len(missing) > 0 {
		return nil, errorz.ErrInvalid{}.New("depositHandler.IsValid; a deposit is missing")
	}
	if len(spent) > 0 {
		return nil, errorz.ErrInvalid{}.New("depositHandler.IsValid; a deposit is already spent")
	}
	return found, nil
}

// Add will add a deposit to the handler
func (dp *Handler) Add(txn *badger.Txn, chainID uint32, utxoID []byte, biValue *big.Int, owner *objs.Owner) error {
	value, err := new(uint256.Uint256).FromBigInt(biValue)
	if err != nil {
		utils.DebugTrace(dp.logger, err)
		return err
	}
	utxoID = utils.CopySlice(utxoID)
	utxoID = utils.ForceSliceToLength(utxoID, constants.HashLen)
	spent, err := dp.IsSpent(txn, utxoID)
	if err != nil {
		utils.DebugTrace(dp.logger, err)
		return err
	}
	if spent {
		return errorz.ErrInvalid{}.New("depositHandler.Add; a deposit is already spent")
	}
	n2 := utils.CopySlice(utxoID)
	vso := &objs.ValueStoreOwner{}
	err = vso.NewFromOwner(owner)
	if err != nil {
		utils.DebugTrace(dp.logger, err)
		return err
	}
	vs := &objs.ValueStore{
		VSPreImage: &objs.VSPreImage{
			TXOutIdx: constants.MaxUint32,
			Value:    value,
			ChainID:  chainID,
			Owner:    vso,
			Fee:      new(uint256.Uint256).SetZero(),
		},
		TxHash: n2,
	}
	utxo := &objs.TXOut{}
	err = utxo.NewValueStore(vs)
	if err != nil {
		utils.DebugTrace(dp.logger, err)
		return err
	}
	err = dp.valueIndex.Add(txn, utxoID, owner, value)
	if err != nil {
		utils.DebugTrace(dp.logger, err)
		return err
	}
	key := dp.makeKey(utxoID)
	_, err = utils.GetValue(txn, key)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(dp.logger, err)
			return err
		}
	} else {
		return errorz.ErrInvalid{}.New("depositHandler.Add; stale")
	}
	if err := db.SetUTXO(txn, key, utxo); err != nil {
		utils.DebugTrace(dp.logger, err)
		return err
	}
	return nil
}

// Remove will delete all references to a deposit from the Handler
func (dp *Handler) Remove(txn *badger.Txn, utxoID []byte) error {
	err := dp.valueIndex.Drop(txn, utxoID)
	if err != nil {
		utils.DebugTrace(dp.logger, err)
	}
	return nil
}

// GetValueForOwner allows a list of utxoIDs to be returned that are equal or
// greater than the value passed as minValue, and are owned by owner.
func (dp *Handler) GetValueForOwner(txn *badger.Txn, owner *objs.Owner, minValue *uint256.Uint256, maxCount int, lastKey []byte) ([][]byte, *uint256.Uint256, []byte, error) {
	excludeSpent := func(utxoID []byte) (bool, error) {
		return dp.IsSpent(txn, utils.CopySlice(utxoID))
	}
	return dp.valueIndex.GetValueForOwner(txn, owner, minValue, excludeSpent, maxCount, lastKey)
}

// Get returns four values <found>, <missing>, <spent>, <error>
// Found returns those deposits that are both known and unspent.
// Missing returns the utxoIDs of the missing deposits.
// Spent returns those deposits found that have been spent.
// An error indicates a failure of the logic and should halt the main.
func (dp *Handler) Get(txn *badger.Txn, utxoIDs [][]byte) ([]*objs.TXOut, [][]byte, []*objs.TXOut, error) {
	f := []*objs.TXOut{}
	m := [][]byte{}
	s := []*objs.TXOut{}
	for _, utxoID := range utxoIDs {
		found, missing, spent, err := dp.getInternal(txn, utils.CopySlice(utxoID))
		if err != nil {
			utils.DebugTrace(dp.logger, err)
			return nil, nil, nil, err
		}
		switch {
		case len(missing) > 0:
			m = append(m, utils.CopySlice(utxoID))
		case spent != nil:
			s = append(s, spent)
		case found != nil:
			f = append(f, found)
		}
	}
	return f, m, s, nil
}

// getInternal returns four values <found>, <missing>, <spent>, <error>;
// look at Get for more information.
func (dp *Handler) getInternal(txn *badger.Txn, utxoID []byte) (*objs.TXOut, []byte, *objs.TXOut, error) {
	utxoID = utils.CopySlice(utxoID)
	utxoID = utils.ForceSliceToLength(utxoID, constants.HashLen)
	n2 := utils.CopySlice(utxoID)
	key := dp.makeKey(n2)
	utxo, err := db.GetUTXO(txn, key)
	if err != nil {
		if err != badger.ErrKeyNotFound {
			utils.DebugTrace(dp.logger, err)
			return nil, nil, nil, err
		}
		return nil, utxoID, nil, nil
	}
	spent, err := dp.IsSpent(txn, utxoID)
	if err != nil {
		utils.DebugTrace(dp.logger, err)
		return nil, nil, nil, err
	}
	if spent {
		return nil, nil, utxo, nil
	}
	return utxo, nil, nil, nil
}

func (dp *Handler) makeKey(utxoID []byte) []byte {
	key := dbprefix.PrefixDeposit()
	key = append(key, utils.CopySlice(utxoID)...)
	return key
}
