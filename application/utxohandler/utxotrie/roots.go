package utxotrie

import (
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/constants/dbprefix"
	"github.com/MadBase/MadNet/utils"
	"github.com/dgraph-io/badger/v2"
)

func SetCurrentStateRoot(txn *badger.Txn, root []byte) error {
	return utils.SetValue(txn, dbprefix.PrefixCurrentStateRoot(), root)
}

func GetCurrentStateRoot(txn *badger.Txn) ([]byte, error) {
	root, err := utils.GetValue(txn, dbprefix.PrefixCurrentStateRoot())
	if err != nil {
		if err != badger.ErrKeyNotFound {
			return nil, err
		}
		return make([]byte, constants.HashLen), nil
	}
	return root, nil
}

func SetPendingStateRoot(txn *badger.Txn, root []byte) error {
	return utils.SetValue(txn, dbprefix.PrefixPendingStateRoot(), root)
}

func GetPendingStateRoot(txn *badger.Txn) ([]byte, error) {
	return utils.GetValue(txn, dbprefix.PrefixPendingStateRoot())
}

func SetCanonicalStateRoot(txn *badger.Txn, root []byte) error {
	return utils.SetValue(txn, dbprefix.PrefixCanonicalStateRoot(), root)
}

func GetCanonicalStateRoot(txn *badger.Txn) ([]byte, error) {
	return utils.GetValue(txn, dbprefix.PrefixCanonicalStateRoot())
}
