package objs

import (
	"github.com/alicenet/alicenet/application/wrapper"
	"github.com/alicenet/alicenet/constants"
)

// TxVec is a vector of transactions Tx
type TxVec []*Tx

// MarshalBinary takes the TxVec object and returns the canonical
// byte slice
func (txv TxVec) MarshalBinary() ([][]byte, error) {
	out := [][]byte{}
	for i := 0; i < len(txv); i++ {
		o, err := txv[i].MarshalBinary()
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, nil
}

// ValidateUnique checks that all inputs and outputs are unique
func (txv TxVec) ValidateUnique(exclSet map[string]bool) (map[string]bool, error) {
	if exclSet == nil {
		exclSet = make(map[string]bool)
	}
	var err error
	for i := 0; i < len(txv); i++ {
		exclSet, err = txv[i].ValidateUnique(exclSet)
		if err != nil {
			return exclSet, err
		}
	}
	return exclSet, nil
}

// ValidateDataStoreIndexes ...
func (txv TxVec) ValidateDataStoreIndexes(exclSet map[string]bool) (map[string]bool, error) {
	if exclSet == nil {
		exclSet = make(map[string]bool)
	}
	var err error
	for i := 0; i < len(txv); i++ {
		exclSet, err = txv[i].ValidateDataStoreIndexes(exclSet)
		if err != nil {
			return exclSet, err
		}
	}
	return exclSet, nil
}

// ValidateChainID validates the chain ID in all transactions
func (txv TxVec) ValidateChainID(chainID uint32) error {
	for i := 0; i < len(txv); i++ {
		err := txv[i].ValidateChainID(chainID)
		if err != nil {
			return err
		}
	}
	return nil
}

// PreValidateApplyState ...
func (txv TxVec) PreValidateApplyState(chainID uint32) error {
	err := txv.ValidateChainID(chainID)
	if err != nil {
		return err
	}
	return nil
}

// Validate ...
func (txv TxVec) Validate(currentHeight uint32, consumedUTXOs Vout, storage *wrapper.Storage) error {
	set := make(map[string]bool)
	voutMap := make(map[[constants.HashLen]byte]*TXOut)
	var key [constants.HashLen]byte
	for i := 0; i < len(consumedUTXOs); i++ {
		utxoID, err := consumedUTXOs[i].UTXOID()
		if err != nil {
			return err
		}
		copy(key[:], utxoID)
		voutMap[key] = consumedUTXOs[i]
	}
	for i := 0; i < len(txv); i++ {
		utxoIDs, err := txv[i].ConsumedUTXOID()
		if err != nil {
			return err
		}
		utxoSet := make([]*TXOut, len(utxoIDs))
		for j := 0; j < len(utxoIDs); j++ {
			copy(key[:], utxoIDs[j])
			utxoSet[j] = voutMap[key]
		}
		if set, err = txv[i].Validate(set, currentHeight, utxoSet, storage); err != nil {
			return err
		}
	}
	return nil
}

// ConsumedUTXOID returns the list of consumed UTXOIDs in TxVec
func (txv TxVec) ConsumedUTXOID() ([][]byte, error) {
	consumed := [][]byte{}
	for i := 0; i < len(txv); i++ {
		c, err := txv[i].ConsumedUTXOID()
		if err != nil {
			return nil, err
		}
		consumed = append(consumed, c...)
	}
	return consumed, nil
}

// ConsumedPreHash returns the list of consumed PreHashs from TxVec
func (txv TxVec) ConsumedPreHash() ([][]byte, error) {
	consumed := [][]byte{}
	for i := 0; i < len(txv); i++ {
		c, err := txv[i].ConsumedPreHash()
		if err != nil {
			return nil, err
		}
		consumed = append(consumed, c...)
	}
	return consumed, nil
}

// ConsumedTxIns returns list of TxIns consumed by TxVec
func (txv TxVec) ConsumedTxIns() (Vin, error) {
	consumed := Vin{}
	for i := 0; i < len(txv); i++ {
		consumed = append(consumed, txv[i].Vin...)
	}
	return consumed, nil
}

// TxHash returns the list of TxHashs from TxVec
func (txv TxVec) TxHash() ([][]byte, error) {
	txhashs := [][]byte{}
	for i := 0; i < len(txv); i++ {
		c, err := txv[i].TxHash()
		if err != nil {
			return nil, err
		}
		txhashs = append(txhashs, c)
	}
	return txhashs, nil
}

// ConsumedIsDeposit returns list of bools to specify which Txs in txv
// are deposits
func (txv TxVec) ConsumedIsDeposit() []bool {
	consumed := []bool{}
	for i := 0; i < len(txv); i++ {
		c := txv[i].ConsumedIsDeposit()
		consumed = append(consumed, c...)
	}
	return consumed
}

// GeneratedUTXOs returns list of TXOuts from Txs in txv
func (txv TxVec) GeneratedUTXOs() (Vout, error) {
	gen := Vout{}
	for i := 0; i < len(txv); i++ {
		c := txv[i].Vout
		gen = append(gen, c...)
	}
	return gen, nil
}

// GeneratedPreHash returns list of PreHashs for Tx in txv
func (txv TxVec) GeneratedPreHash() ([][]byte, error) {
	gen := [][]byte{}
	for i := 0; i < len(txv); i++ {
		c, err := txv[i].GeneratedPreHash()
		if err != nil {
			return nil, err
		}
		gen = append(gen, c...)
	}
	return gen, nil
}

// GeneratedUTXOID returns list of UTXOIDs for Tx in txv
func (txv TxVec) GeneratedUTXOID() ([][]byte, error) {
	gen := [][]byte{}
	for i := 0; i < len(txv); i++ {
		c, err := txv[i].GeneratedUTXOID()
		if err != nil {
			return nil, err
		}
		gen = append(gen, c...)
	}
	return gen, nil
}

// ConsumedPreHashOnlyDeposits returns list of PreHashs from txv only for
// transactions which are deposits
func (txv TxVec) ConsumedPreHashOnlyDeposits() ([][]byte, error) {
	consumed := [][]byte{}
	for i := 0; i < len(txv); i++ {
		tx := txv[i]
		for j := 0; j < len(tx.Vin); j++ {
			isDep := tx.Vin[j].IsDeposit()
			if isDep {
				preHash, err := tx.Vin[j].PreHash()
				if err != nil {
					return nil, err
				}
				consumed = append(consumed, preHash)
			}
		}
	}
	return consumed, nil
}

// ConsumedUTXOIDOnlyDeposits returns list of UTXOIDs from txv only for
// transactions which are deposits
func (txv TxVec) ConsumedUTXOIDOnlyDeposits() ([][]byte, error) {
	consumed := [][]byte{}
	for i := 0; i < len(txv); i++ {
		tx := txv[i]
		for j := 0; j < len(tx.Vin); j++ {
			isDep := tx.Vin[j].IsDeposit()
			if isDep {
				utxoID, err := tx.Vin[j].UTXOID()
				if err != nil {
					return nil, err
				}
				consumed = append(consumed, utxoID)
			}
		}
	}
	return consumed, nil
}

// ConsumedUTXOIDNoDeposits returns list of UTXOIDs from txv only for
// transactions which are not deposits
func (txv TxVec) ConsumedUTXOIDNoDeposits() ([][]byte, error) {
	consumed := [][]byte{}
	for i := 0; i < len(txv); i++ {
		tx := txv[i]
		for j := 0; j < len(tx.Vin); j++ {
			isDep := tx.Vin[j].IsDeposit()
			if !isDep {
				utxoID, err := tx.Vin[j].UTXOID()
				if err != nil {
					return nil, err
				}
				consumed = append(consumed, utxoID)
			}
		}
	}
	return consumed, nil
}

// PreValidatePending ...
func (txv TxVec) PreValidatePending(chainID uint32) error {
	for i := 0; i < len(txv); i++ {
		if err := txv[i].PreValidatePending(chainID); err != nil {
			return err
		}
	}
	return nil
}

// PostValidatePending ...
func (txv TxVec) PostValidatePending(currentHeight uint32, consumedUTXOs Vout, storage *wrapper.Storage) error {
	for i := 0; i < len(txv); i++ {
		if err := txv[i].PostValidatePending(currentHeight, consumedUTXOs, storage); err != nil {
			return err
		}
	}
	return nil
}
