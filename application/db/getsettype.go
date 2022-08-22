package db

import (
	"github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/utils"
	"github.com/dgraph-io/badger/v2"
)

// SetUTXO will set a UTXO in the database.
func SetUTXO(txn *badger.Txn, key []byte, v *objs.TXOut) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

// GetUTXO will set a UTXO in the database.
func GetUTXO(txn *badger.Txn, key []byte) (*objs.TXOut, error) {
	v, err := utils.GetValue(txn, key)
	if err != nil {
		return nil, err
	}
	vv := &objs.TXOut{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}

// SetTx will set a Tx in the database.
func SetTx(txn *badger.Txn, key []byte, v *objs.Tx) error {
	vv, err := v.MarshalBinary()
	if err != nil {
		return err
	}
	return utils.SetValue(txn, key, vv)
}

// GetTx will set a Tx in the database.
func GetTx(txn *badger.Txn, key []byte) (*objs.Tx, error) {
	v, err := utils.GetValue(txn, key)
	if err != nil {
		return nil, err
	}
	vv := &objs.Tx{}
	err = vv.UnmarshalBinary(v)
	if err != nil {
		return nil, err
	}
	return vv, nil
}
