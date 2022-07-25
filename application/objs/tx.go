package objs

import (
	"bytes"
	"fmt"

	"github.com/alicenet/alicenet/errorz"

	capnp "github.com/MadBase/go-capnproto2/v2"
	mdefs "github.com/alicenet/alicenet/application/objs/capn"
	"github.com/alicenet/alicenet/application/objs/tx"
	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/application/wrapper"
	trie "github.com/alicenet/alicenet/badgerTrie"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/utils"
)

var _ interfaces.Transaction = (*Tx)(nil)

// Tx is a transaction object
type Tx struct {
	Vin  Vin
	Vout Vout
	Fee  *uint256.Uint256
	// not part of serialized object below this line
	txHash []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// Tx object
func (b *Tx) UnmarshalBinary(data []byte) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.UnmarshalBinary: tx not initialized")
	}
	if len(data) == 0 {
		return errorz.ErrInvalid{}.New("tx.UnmarshalBinary: len(data) == 0")
	}
	if len(data) > constants.MaxTxSize {
		return errorz.ErrInvalid{}.New("tx.UnmarshalBinary: len(data) > MaxTxSize")
	}
	if len(data) == 0 {
		return errorz.ErrInvalid{}.New("tx.unmarshalBinary: len(data) == 0")
	}
	if len(data) > constants.MaxTxSize {
		return errorz.ErrInvalid{}.New("tx.unmarshalBinary: len(data) > MaxTxSize")
	}
	bc, err := tx.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the Tx object and returns the canonical
// byte slice
func (b *Tx) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.MarshalBinary: tx not initialized")
	}
	if b.Fee == nil {
		return nil, errorz.ErrInvalid{}.New("tx.MarshalBinary: tx.fee not initialized")
	}
	if len(b.Vin) > constants.MaxTxVectorLength {
		return nil, errorz.ErrInvalid{}.New("tx.MarshalBinary: len(tx.vin) > MaxTxVectorLength")
	}
	if len(b.Vout) > constants.MaxTxVectorLength {
		return nil, errorz.ErrInvalid{}.New("tx.MarshalBinary: len(tx.vout) > MaxTxVectorLength")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	txBytes, err := tx.Marshal(bc)
	if err != nil {
		return nil, err
	}
	if len(txBytes) == 0 {
		return nil, errorz.ErrInvalid{}.New("tx.MarshalBinary: len(txBytes) == 0")
	}
	if len(txBytes) > constants.MaxTxSize {
		return nil, errorz.ErrInvalid{}.New("tx.MarshalBinary: len(txBytes) > MaxTxSize")
	}
	return txBytes, nil
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *Tx) UnmarshalCapn(bc mdefs.Tx) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.UnmarshalCapn: tx not initialized")
	}
	if err := tx.Validate(bc); err != nil {
		return err
	}
	invec := []*TXIn{}
	vin, err := bc.Vin()
	if err != nil {
		return err
	}
	for i := 0; i < vin.Len(); i++ {
		mobj := vin.At(i)
		txin := &TXIn{}
		if err := txin.UnmarshalCapn(mobj); err != nil {
			return err
		}
		invec = append(invec, txin)
	}

	outvec := []*TXOut{}
	vout, err := bc.Vout()
	if err != nil {
		return err
	}
	for i := 0; i < vout.Len(); i++ {
		mobj := vout.At(i)
		txout := &TXOut{}
		if err := txout.UnmarshalCapn(mobj); err != nil {
			return err
		}
		outvec = append(outvec, txout)
	}
	b.Vin = invec
	b.Vout = outvec
	u32array := [8]uint32{}
	u32array[0] = bc.Fee0()
	u32array[1] = bc.Fee1()
	u32array[2] = bc.Fee2()
	u32array[3] = bc.Fee3()
	u32array[4] = bc.Fee4()
	u32array[5] = bc.Fee5()
	u32array[6] = bc.Fee6()
	u32array[7] = bc.Fee7()
	fObj := &uint256.Uint256{}
	err = fObj.FromUint32Array(u32array)
	if err != nil {
		return err
	}
	b.Fee = fObj
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *Tx) MarshalCapn(seg *capnp.Segment) (mdefs.Tx, error) {
	if b == nil {
		return mdefs.Tx{}, errorz.ErrInvalid{}.New("tx.MarshalCapn: tx not initialized")
	}
	var bc mdefs.Tx
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootTx(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewTx(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	seg = bc.Struct.Segment()
	vin, err := bc.NewVin(int32(len(b.Vin)))
	if err != nil {
		return bc, err
	}
	for i := 0; i < len(b.Vin); i++ {
		txin := b.Vin[i]
		mobj, err := txin.MarshalCapn(seg)
		if err != nil {
			return bc, err
		}
		if err := vin.Set(i, mobj); err != nil {
			return bc, err
		}
	}
	vout, err := bc.NewVout(int32(len(b.Vout)))
	if err != nil {
		return bc, err
	}
	for i := 0; i < len(b.Vout); i++ {
		txout := b.Vout[i]
		mobj, err := txout.MarshalCapn(seg)
		if err != nil {
			return bc, err
		}
		if err := vout.Set(i, mobj); err != nil {
			return bc, err
		}
	}
	if err := bc.SetVin(vin); err != nil {
		return bc, err
	}
	if err := bc.SetVout(vout); err != nil {
		return bc, err
	}
	u32array, err := b.Fee.ToUint32Array()
	if err != nil {
		return bc, err
	}
	bc.SetFee0(u32array[0])
	bc.SetFee1(u32array[1])
	bc.SetFee2(u32array[2])
	bc.SetFee3(u32array[3])
	bc.SetFee4(u32array[4])
	bc.SetFee5(u32array[5])
	bc.SetFee6(u32array[6])
	bc.SetFee7(u32array[7])
	return bc, nil
}

// ValidateUnique checks that all inputs and outputs are unique
func (b *Tx) ValidateUnique(opset map[string]bool) (map[string]bool, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.ValidateUnique: tx not initialized")
	}
	if opset == nil {
		opset = make(map[string]bool)
	}
	for i := 0; i < len(b.Vin); i++ {
		hsh, err := b.Vin[i].UTXOID()
		if err != nil {
			return nil, err
		}
		if !opset[string(hsh)] {
			opset[string(hsh)] = true
			continue
		}
		return nil, errorz.ErrInvalid{}.New("tx.ValidateUnique: duplicate input")
	}
	for i := 0; i < len(b.Vout); i++ {
		hsh, err := b.Vout[i].UTXOID()
		if err != nil {
			return nil, err
		}
		if !opset[string(hsh)] {
			opset[string(hsh)] = true
			continue
		}
		return nil, errorz.ErrInvalid{}.New("tx.ValidateUnique: duplicate output")
	}
	return opset, nil
}

// ValidateDataStoreIndexes ensures there are no duplicate output indices
// for DataStore objects
func (b *Tx) ValidateDataStoreIndexes(opset map[string]bool) (map[string]bool, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.ValidateDataStoreIndexes: tx not initialized")
	}
	if opset == nil {
		opset = make(map[string]bool)
	}
	for _, utxo := range b.Vout {
		if utxo.HasDataStore() {
			ds, err := utxo.DataStore()
			if err != nil {
				return nil, err
			}
			index, err := ds.Index()
			if err != nil {
				return nil, err
			}
			owner, err := utxo.GenericOwner()
			if err != nil {
				return nil, err
			}
			ownerBytes, err := owner.MarshalBinary()
			if err != nil {
				return nil, err
			}
			tmp := []byte{}
			tmp = append(tmp, ownerBytes...)
			tmp = append(tmp, index...)
			hsh := crypto.Hasher(tmp)
			if !opset[string(hsh)] {
				opset[string(hsh)] = true
				continue
			}
			return nil, errorz.ErrInvalid{}.New("tx.ValidateDataStoreIndexes: duplicate output index for datastore")
		}
	}
	return opset, nil
}

// ValidateTxHash validates the txHash is correct on all objects
func (b *Tx) ValidateTxHash() error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.ValidateTxHash: tx not initialized")
	}
	if b.txHash != nil {
		return nil
	}
	txHash, err := b.TxHash()
	if err != nil {
		return err
	}
	for _, txIn := range b.Vin {
		txInTxHash, err := txIn.TxHash()
		if err != nil {
			return err
		}
		if !bytes.Equal(txInTxHash, txHash) {
			return errorz.ErrInvalid{}.New("tx.ValidateTxHash: wrong txHash (vin)")
		}
	}
	for _, txOut := range b.Vout {
		hsh, err := txOut.TxHash()
		if err != nil {
			return err
		}
		if !bytes.Equal(hsh, txHash) {
			return errorz.ErrInvalid{}.New("tx.ValidateTxHash: wrong txHash (vout)")
		}
	}
	return nil
}

// TxHash calculates the TxHash of the transaction
func (b *Tx) TxHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.TxHash: tx not initialized")
	}
	if b.Fee == nil {
		return nil, errorz.ErrInvalid{}.New("tx.TxHash: tx.fee not initialized")
	}
	if b.txHash != nil {
		return utils.CopySlice(b.txHash), nil
	}
	if err := b.Vout.SetTxOutIdx(); err != nil {
		return nil, err
	}
	keys := [][]byte{}
	values := [][]byte{}
	for _, txIn := range b.Vin {
		id, err := txIn.UTXOID()
		if err != nil {
			return nil, err
		}
		keys = append(keys, id)
		hsh, err := txIn.PreHash()
		if err != nil {
			return nil, err
		}
		values = append(values, hsh)
	}
	for idx, txOut := range b.Vout {
		hsh, err := txOut.PreHash()
		if err != nil {
			return nil, err
		}
		id := MakeUTXOID(utils.CopySlice(hsh), uint32(idx))
		keys = append(keys, id)
		values = append(values, hsh)
	}
	// new in memory smt
	smt := trie.NewMemoryTrie()
	// smt update
	keysSorted, valuesSorted, err := utils.SortKVs(keys, values)
	if err != nil {
		return nil, err
	}
	if len(keysSorted) == 0 && len(valuesSorted) == 0 {
		rootHash := crypto.Hasher([][]byte{}...)
		b.txHash = rootHash
		return utils.CopySlice(b.txHash), nil
	}
	rootHash, err := smt.Update(keysSorted, valuesSorted)
	if err != nil {
		return nil, err
	}
	b.txHash = rootHash
	return utils.CopySlice(b.txHash), nil
}

// SetTxHash calculates the TxHash and sets it on all UTXOs and TXIns
func (b *Tx) SetTxHash() error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.SetTxHash: tx not initialized")
	}
	if b.Fee == nil {
		return errorz.ErrInvalid{}.New("tx.SetTxHash: tx.fee not initialized")
	}
	txHash, err := b.TxHash()
	if err != nil {
		return err
	}
	for _, txIn := range b.Vin {
		if err := txIn.SetTxHash(txHash); err != nil {
			return err
		}
	}
	for _, txOut := range b.Vout {
		if err := txOut.SetTxHash(txHash); err != nil {
			return err
		}
	}
	return nil
}

// ConsumedPreHash returns the list of PreHashs from Vin
func (b *Tx) ConsumedPreHash() ([][]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.ConsumedPreHash: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return nil, errorz.ErrInvalid{}.New("tx.ConsumedPreHash: tx.vin not initialized")
	}
	return b.Vin.PreHash()
}

// ConsumedUTXOID returns the list of UTXOIDs from Vin
func (b *Tx) ConsumedUTXOID() ([][]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.ConsumedUTXOID: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return nil, errorz.ErrInvalid{}.New("tx.ConsumedUTXOID: tx.vin not initialized")
	}
	return b.Vin.UTXOID()
}

// ConsumedIsDeposit returns the list of IsDeposit bools from Vin
func (b *Tx) ConsumedIsDeposit() []bool {
	return b.Vin.IsDeposit()
}

// GeneratedUTXOID returns the list of UTXOIDs from Vout
func (b *Tx) GeneratedUTXOID() ([][]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.GeneratedUTXOID: tx not initialized")
	}
	if len(b.Vout) == 0 {
		return nil, errorz.ErrInvalid{}.New("tx.GeneratedUTXOID: tx.vout not initialized")
	}
	return b.Vout.UTXOID()
}

// GeneratedPreHash returns the list of PreHashs from Vout
func (b *Tx) GeneratedPreHash() ([][]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.GeneratedPreHash: tx not initialized")
	}
	if len(b.Vout) == 0 {
		return nil, errorz.ErrInvalid{}.New("tx.GeneratedPreHash: tx.vout not initialized")
	}
	return b.Vout.PreHash()
}

// ValidateSignature validates the signatures of the objects
func (b *Tx) ValidateSignature(currentHeight uint32, refUTXOs Vout) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.ValidateSignature: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidateSignature: tx.vin not initialized")
	}
	return refUTXOs.ValidateSignature(currentHeight, b.Vin)
}

// ValidatePreSignature validates the presignatures of the objects
func (b *Tx) ValidatePreSignature() error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.ValidatePreSignature: tx not initialized")
	}
	if len(b.Vout) == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidatePreSignature: tx.vout not initialized")
	}
	return b.Vout.ValidatePreSignature()
}

// ValidateFees validates the fees of the object.
// currentHeight and refUTXOs are needed to verify if we have a cleanup tx.
func (b *Tx) ValidateFees(currentHeight uint32, refUTXOs Vout, storage *wrapper.Storage) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.ValidateFees: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidateFees: tx.vin not initialized")
	}
	if len(b.Vout) == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidateFees: tx.vout not initialized")
	}
	if b.Fee == nil {
		return errorz.ErrInvalid{}.New("tx.ValidateFees: tx.fee not initialized")
	}
	if b.IsCleanupTx(currentHeight, refUTXOs) {
		// Tx is a valid Cleanup Tx, so we do not worry about fees;
		// the specific cleanup tx fees are already performed in this call.
		return nil
	}
	if err := b.Vout.ValidateFees(storage); err != nil {
		return err
	}
	// Ensure FeeCostRatio is above minimum
	minTxFeeCostRatio, err := storage.GetMinTxFeeCostRatio()
	if err != nil {
		return err
	}
	// isCleanup is false because it is not a cleanup tx
	feeCostRatio, err := b.ScaledFeeCostRatio(false)
	if err != nil {
		return err
	}
	if feeCostRatio.Lt(minTxFeeCostRatio) {
		return errorz.ErrInvalid{}.New("tx.ValidateFees; feeCostRatio below minimum")
	}
	return nil
}

// Cost computes the total transaction complexity
func (b *Tx) cost() (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.cost: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return nil, errorz.ErrInvalid{}.New("tx.cost: tx.vin not initialized")
	}
	if len(b.Vout) == 0 {
		return nil, errorz.ErrInvalid{}.New("tx.cost: tx.vout not initialized")
	}
	costSize, err := b.costSize()
	if err != nil {
		return nil, err
	}
	costComputation, err := b.costComputation()
	if err != nil {
		return nil, err
	}
	costTotal, err := new(uint256.Uint256).Add(costSize, costComputation)
	if err != nil {
		return nil, err
	}
	if costTotal.IsZero() {
		return nil, errorz.ErrInvalid{}.New("tx.cost: cost is zero")
	}
	return costTotal, nil
}

// costSize computes the total transaction complexity due to size (bytes)
func (b *Tx) costSize() (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.costSize: tx not initialized")
	}
	txBytes, err := b.MarshalBinary()
	if err != nil {
		return nil, err
	}
	size := len(txBytes)
	costSize := &uint256.Uint256{}
	costSize.FromUint64(uint64(size))
	return costSize, nil
}

// costComputation computes the total transaction complexity due to computation.
// TODO: update this appropriately.
func (b *Tx) costComputation() (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.costComputation: tx not initialized")
	}
	costComputation := uint256.Zero()
	return costComputation, nil
}

// ScaledFeeCostRatio returns the fee cost ratio as a uint256
func (b *Tx) ScaledFeeCostRatio(isCleanup bool) (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.ScaledFeeCostRatio: tx not initialized")
	}
	if b.Fee == nil {
		return nil, errorz.ErrInvalid{}.New("tx.ScaledFeeCostRatio: tx.Fee not initialized")
	}
	cost, err := b.cost()
	if err != nil {
		return nil, err
	}
	if isCleanup {
		if !b.Fee.IsZero() {
			return nil, errorz.ErrInvalid{}.New("tx.ScaledFeeCostRatio: is cleanup tx but tx.Fee is nonzero")
		}
		feeCostRatio := uint256.Max()
		return feeCostRatio, nil
	}
	scaleFactor := uint256.TwoPower64()
	maxFee := uint256.TwoPower128()
	feeCopy := b.Fee.Clone()
	// Ensure the fee is not too large
	if feeCopy.Gt(maxFee) {
		err := feeCopy.Set(maxFee)
		if err != nil {
			return nil, err
		}
	}
	scaledFee, err := new(uint256.Uint256).Mul(feeCopy, scaleFactor)
	if err != nil {
		return nil, err
	}
	feeCostRatio, err := new(uint256.Uint256).Div(scaledFee, cost)
	if err != nil {
		return nil, err
	}
	return feeCostRatio, nil
}

// ValidateEqualVinVout checks the following
// calc sum on inputs from utxos and currentHeight
// sum inputs must equal sum outputs plus fee.
// Note well: This function call checks the values concerning MadBytes.
// It also validates logic related the ERCToken object
// by calling ValidateERCTokens.
func (b *Tx) ValidateEqualVinVout(currentHeight uint32, refUTXOs Vout) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.ValidateEqualVinVout: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidateEqualVinVout: tx.vin not initialized")
	}
	if len(b.Vout) == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidateEqualVinVout: tx.vout not initialized")
	}
	if b.Fee == nil {
		return errorz.ErrInvalid{}.New("tx.ValidateEqualVinVout: tx.fee not initialized")
	}
	minBH, err := b.CannotBeMinedUntil()
	if err != nil {
		return err
	}
	if minBH > currentHeight {
		// We cannot mine before the future;
		// to calculate the correct future value, we must look at the height
		// we will mine it in the future
		currentHeight = minBH
	}
	valueIn, err := refUTXOs.RemainingValue(currentHeight)
	if err != nil {
		return err
	}
	valueOut, err := b.Vout.ValuePlusFee()
	if err != nil {
		return err
	}
	valueOutPlusFee, err := new(uint256.Uint256).Add(valueOut, b.Fee)
	if err != nil {
		return err
	}
	if !valueOutPlusFee.Eq(valueIn) {
		return errorz.ErrInvalid{}.New(fmt.Sprintf("tx.ValidateEqualVinVout: input value does not match output value: IN:%v  vs  OUT+FEE:%v", valueIn, valueOutPlusFee))
	}
	return nil
}

// ValidateChainID validates that all elements have the correct ChainID
func (b *Tx) ValidateChainID(chainID uint32) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.ValidateChainID: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidateChainID: tx.vin not initialized")
	}
	if len(b.Vout) == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidateChainID: tx.vout not initialized")
	}
	if chainID == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidateChainID: chainID invalid; cannot be 0")
	}
	for _, inp := range b.Vin {
		inpCid, err := inp.ChainID()
		if err != nil {
			return err
		}
		if inpCid != chainID {
			return errorz.ErrInvalid{}.New("tx.ValidateChainID: invalid chainID; bad chain ID (vin)")
		}
	}
	for _, outp := range b.Vout {
		outpCid, err := outp.ChainID()
		if err != nil {
			return err
		}
		if outpCid != chainID {
			return errorz.ErrInvalid{}.New("tx.ValidateChainID: invalid chainID; bad chain ID (vout)")
		}
	}
	return nil
}

// CannotBeMinedUntil ...
func (b *Tx) CannotBeMinedUntil() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("tx.CannotBeMinedUntil: tx not initialized")
	}
	if len(b.Vout) == 0 {
		return 0, errorz.ErrInvalid{}.New("tx.CannotBeMinedUntil: tx.vout not initialized")
	}
	maxBH := uint32(1)
	for _, utxo := range b.Vout {
		mbh, err := utxo.CannotBeMinedBeforeHeight()
		if err != nil {
			return 0, err
		}
		if mbh > maxBH {
			maxBH = mbh
		}
	}
	return maxBH, nil
}

// ValidateIssuedAtForMining ...
func (b *Tx) ValidateIssuedAtForMining(currentHeight uint32) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.ValidateIssuedAtForMining: tx not initialized")
	}
	if len(b.Vout) == 0 {
		return errorz.ErrInvalid{}.New("tx.ValidateIssuedAtForMining: tx.vout not initialized")
	}
	hmap := make(map[uint32]bool)
	for _, utxo := range b.Vout {
		mbh, err := utxo.MustBeMinedBeforeHeight()
		if err != nil {
			return err
		}
		if mbh != constants.MaxUint32 {
			hmap[mbh] = true
		}
	}
	if len(hmap) == 0 {
		return nil
	}
	if len(hmap) > 1 {
		return errorz.ErrInvalid{}.New("tx.ValidateIssuedAtForMining: conflicting IssuedAt")
	}
	mbh := uint32(0)
	for k := range hmap {
		mbh = k
		break
	}
	if utils.Epoch(mbh) != utils.Epoch(currentHeight) {
		return errorz.ErrInvalid{}.New("tx.ValidateIssuedAtForMining: mining out of epoch")
	}
	return nil
}

// EpochOfExpirationForMining ...
func (b *Tx) EpochOfExpirationForMining() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("tx.EpochOfExpirationForMining: tx not initialized")
	}
	if len(b.Vout) == 0 {
		return 0, errorz.ErrInvalid{}.New("tx.EpochOfExpirationForMining: tx.vout not initialized")
	}
	hmap := make(map[uint32]bool)
	for _, utxo := range b.Vout {
		mbh, err := utxo.MustBeMinedBeforeHeight()
		if err != nil {
			return 0, err
		}
		if mbh != constants.MaxUint32 {
			hmap[mbh] = true
		}
	}
	if len(hmap) == 0 {
		return constants.MaxUint32, nil
	}
	if len(hmap) > 1 {
		return 0, errorz.ErrInvalid{}.New("tx.EpochOfExpirationForMining: conflicting IssuedAt")
	}
	mbh := uint32(0)
	for k := range hmap {
		mbh = k
		break
	}
	return utils.Epoch(mbh), nil
}

// Validate ...
func (b *Tx) Validate(set map[string]bool, currentHeight uint32, consumedUTXOs Vout, storage *wrapper.Storage) (map[string]bool, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("tx.Validate: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return nil, errorz.ErrInvalid{}.New("tx.Validate: tx.vin not initialized")
	}
	if len(b.Vout) == 0 {
		return nil, errorz.ErrInvalid{}.New("tx.Validate: tx.vout not initialized")
	}
	if b.Fee == nil {
		return nil, errorz.ErrInvalid{}.New("tx.Validate: tx.fee not initialized")
	}
	if err := b.Vout.ValidateTxOutIdx(); err != nil {
		return nil, err
	}
	set, err := b.ValidateDataStoreIndexes(set)
	if err != nil {
		return nil, err
	}
	set, err = b.ValidateUnique(set)
	if err != nil {
		return nil, err
	}
	err = b.ValidateEqualVinVout(currentHeight, consumedUTXOs)
	if err != nil {
		return nil, err
	}
	err = b.ValidateFees(currentHeight, consumedUTXOs, storage)
	if err != nil {
		return nil, err
	}
	err = b.ValidateTxHash()
	if err != nil {
		return nil, err
	}
	return set, nil
}

// PreValidatePending performs the initial check of a transaction
// before it is added to the pending transaction pool.
// This includes all of the basic validation logic that *all* transactions
// must minimally pass to be valid.
// Other important validation logic is included in PostValidatePending.
func (b *Tx) PreValidatePending(chainID uint32) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.PreValidatePending: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return errorz.ErrInvalid{}.New("tx.PreValidatePending: tx.vin not initialized")
	}
	if len(b.Vout) == 0 {
		return errorz.ErrInvalid{}.New("tx.PreValidatePending: tx.vout not initialized")
	}
	if b.Fee == nil {
		return errorz.ErrInvalid{}.New("tx.PreValidatePending: tx.fee not initialized")
	}
	err := b.ValidateChainID(chainID)
	if err != nil {
		return err
	}
	_, err = b.ValidateUnique(nil)
	if err != nil {
		return err
	}
	err = b.ValidateTxHash()
	if err != nil {
		return err
	}
	err = b.ValidatePreSignature()
	if err != nil {
		return err
	}
	return nil
}

// PostValidatePending performs the validation logic which occurs before
// adding the transaction to the pending transaction pool.
func (b *Tx) PostValidatePending(currentHeight uint32, consumedUTXOs Vout, storage *wrapper.Storage) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("tx.PostValidatePending: tx not initialized")
	}
	if len(b.Vin) == 0 {
		return errorz.ErrInvalid{}.New("tx.PostValidatePending: tx.vin not initialized")
	}
	if len(b.Vout) == 0 {
		return errorz.ErrInvalid{}.New("tx.PostValidatePending: tx.vout not initialized")
	}
	if b.Fee == nil {
		return errorz.ErrInvalid{}.New("tx.PostValidatePending: tx.fee not initialized")
	}
	err := b.ValidateEqualVinVout(currentHeight, consumedUTXOs)
	if err != nil {
		return err
	}
	err = b.ValidateFees(currentHeight, consumedUTXOs, storage)
	if err != nil {
		return err
	}
	err = b.ValidateSignature(currentHeight, consumedUTXOs)
	if err != nil {
		return err
	}
	return nil
}

// IsPotentialCleanupTx checks if the tx meets the minimal requriements
// of a cleanup transaction. This potentially reduces total work by
// not requiring currentHeight or referenced utxos.
// There are no false negatives, only potential false positives.
func (b *Tx) IsPotentialCleanupTx() bool {
	if b == nil || b.Fee == nil {
		return false
	}
	// Confirm Fee is zero
	if !b.Fee.IsZero() {
		return false
	}
	// Confirm Vout
	cleanupVout := b.Vout.IsCleanupVout()
	return cleanupVout
}

// IsCleanupTx checks if Tx is a cleanup transaction.
// Cleanup transactions are unique in that there is no associated TxFee
// or ValueStoreFee. The CleaupTx must consist of *expired* DataStores
// with value equal to that in the only ValueStore in Vout.
func (b *Tx) IsCleanupTx(currentHeight uint32, refUTXOs Vout) bool {
	// Confirm Fee is zero and Vout
	if !b.IsPotentialCleanupTx() {
		return false
	}
	// Confirm Vin
	cleanupVin := b.Vin.IsCleanupVin(currentHeight, refUTXOs)
	if !cleanupVin {
		return false
	}
	// Confirm inputs equal outputs
	if err := b.ValidateEqualVinVout(currentHeight, refUTXOs); err != nil {
		return false
	}
	return true
}

// XXXIsTx allows compile time type checking for transaction interfaces
func (b *Tx) XXXIsTx() {}
