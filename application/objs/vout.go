package objs

import (
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/application/wrapper"
	"github.com/alicenet/alicenet/errorz"
)

// Vout is a vector of TXOut objects
type Vout []*TXOut

// ValuePlusFee sums the total value of the UTXOs without any discount
// and including associated fees
func (vout Vout) ValuePlusFee() (*uint256.Uint256, error) {
	sum := uint256.Zero()
	for i := 0; i < len(vout); i++ {
		value, err := vout[i].ValuePlusFee()
		if err != nil {
			return nil, err
		}
		sum, err = sum.Add(sum, value)
		if err != nil {
			return nil, err
		}
	}
	return sum, nil
}

// RemainingValue sums the total value of the UTXOs with discount
func (vout Vout) RemainingValue(currentHeight uint32) (*uint256.Uint256, error) {
	sum := uint256.Zero()
	for i := 0; i < len(vout); i++ {
		value, err := vout[i].RemainingValue(currentHeight)
		if err != nil {
			return nil, err
		}
		sum, err = sum.Add(sum, value)
		if err != nil {
			return nil, err
		}
	}
	return sum, nil
}

// SetTxOutIdx sets the TxOutIdx of each utxo
func (vout Vout) SetTxOutIdx() error {
	for i := 0; i < len(vout); i++ {
		err := vout[i].SetTxOutIdx(uint32(i))
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateTxOutIdx validates the TxOutIdx of each utxo
func (vout Vout) ValidateTxOutIdx() error {
	var txOutIdx uint32
	idxMap := make(map[uint32]bool)
	for i := 0; i < len(vout); i++ {
		utxo := vout[i]
		switch {
		case utxo.HasDataStore():
			ds, _ := utxo.DataStore()
			dsTxOutIdx, err := ds.TxOutIdx()
			if err != nil {
				return err
			}
			txOutIdx = dsTxOutIdx
		case utxo.HasValueStore():
			vs, _ := utxo.ValueStore()
			vsTxOutIdx, err := vs.TxOutIdx()
			if err != nil {
				return err
			}
			txOutIdx = vsTxOutIdx
		case utxo.HasAtomicSwap():
			as, _ := utxo.AtomicSwap()
			asTxOutIdx, err := as.TxOutIdx()
			if err != nil {
				return err
			}
			txOutIdx = asTxOutIdx
		default:
			return errorz.ErrInvalid{}.New("vout.validateTxOutIdx; bad txOutIdx: Invalid Type")
		}
		if idxMap[txOutIdx] {
			return errorz.ErrInvalid{}.New("vout.validateTxOutIdx; duplicate txOutIdx")
		}
		idxMap[txOutIdx] = true
	}
	for i := uint32(0); i < uint32(len(idxMap)); i++ {
		if !idxMap[i] {
			return errorz.ErrInvalid{}.New("vout.validateTxOutIdx; missing tx out index")
		}
	}
	return nil
}

// UTXOID returns the list of UTXOIDs from each TXOut in Vout
func (vout Vout) UTXOID() ([][]byte, error) {
	ids := [][]byte{}
	for i := 0; i < len(vout); i++ {
		id, err := vout[i].UTXOID()
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// PreHash returns the list of PreHashs from each TXOut in Vout
func (vout Vout) PreHash() ([][]byte, error) {
	phs := [][]byte{}
	for i := 0; i < len(vout); i++ {
		ph, err := vout[i].PreHash()
		if err != nil {
			return nil, err
		}
		phs = append(phs, ph)
	}
	return phs, nil
}

// ValidateFees validates the Fee from each TXOut in Vout
func (vout Vout) ValidateFees(storage *wrapper.Storage) error {
	for i := 0; i < len(vout); i++ {
		err := vout[i].ValidateFee(storage)
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidatePreSignature validates the PreSignature from each TXOut in Vout
func (vout Vout) ValidatePreSignature() error {
	for i := 0; i < len(vout); i++ {
		err := vout[i].ValidatePreSignature()
		if err != nil {
			return err
		}
	}
	return nil
}

// ValidateSignature validates the Signature from each TXOut in Vout
func (vout Vout) ValidateSignature(currentHeight uint32, txIn []*TXIn) error {
	if len(txIn) != len(vout) {
		return errorz.ErrInvalid{}.New("vout.validateSignature; mismatched vector lengths")
	}
	for i := 0; i < len(vout); i++ {
		err := vout[i].ValidateSignature(currentHeight, txIn[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// MakeTxIn converts Vout to Vin
func (vout Vout) MakeTxIn() (Vin, error) {
	txIns := Vin{}
	for i := 0; i < len(vout); i++ {
		txin, err := vout[i].MakeTxIn()
		if err != nil {
			return nil, err
		}
		txIns = append(txIns, txin)
	}
	return txIns, nil
}

// IsCleanupVout ensures we have a valid Vout object in Cleanup Tx.
// In this case, Vout must be only one ValueStore with no fee.
func (vout Vout) IsCleanupVout() bool {
	if len(vout) != 1 {
		return false
	}
	// Confirm utxo is ValueStore with no fee
	utxo := vout[0]
	if !utxo.HasValueStore() {
		return false
	}
	vs, err := utxo.ValueStore()
	if err != nil {
		return false
	}
	vsFee, err := vs.Fee()
	if err != nil {
		return false
	}
	if !vsFee.IsZero() {
		return false
	}
	return true
}
