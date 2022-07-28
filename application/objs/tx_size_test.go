package objs

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

func TestTxValueTransferSize1(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 1, valueOut, vsfee)

	// Setup Tx
	vin := []*TXIn{txin1}
	vout := []*TXOut{utxo2}
	refUTXOs := Vout{utxo1}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign object
	err = utxo1.valueStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Tx with 1->1 value transfer")
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxValueTransferSize2(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make inputs and outputs
	totalInputsAndOutputs := 16
	consumedUTXOs := Vout{}
	vin := Vin{}
	vout := Vout{}
	value64 := uint64(constants.MaxUint64)
	for k := 0; k < totalInputsAndOutputs; k++ {
		v := value64 - uint64(k)
		valueIn, err := new(uint256.Uint256).FromUint64(v)
		if err != nil {
			t.Fatal(err)
		}
		// Make consumed utxo
		consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
		valueBytes := utils.MarshalUint64(v)
		indexBytes := utils.MarshalUint64(uint64(k))
		vHash := crypto.Hasher(valueBytes, indexBytes)
		err = consumedUTXO.SetTxHash(vHash)
		if err != nil {
			t.Fatal(err)
		}
		consumedUTXOs = append(consumedUTXOs, consumedUTXO)
		txin, err := consumedUTXO.MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		vin = append(vin, txin)
		// Make corresponding utxo
		valueOut, err := new(uint256.Uint256).Sub(valueIn, vsfee)
		if err != nil {
			t.Fatal(err)
		}
		if k == 0 {
			// Subtract txfee from initial
			_, err = valueOut.Sub(valueOut, minTxFee)
			if err != nil {
				t.Fatal(err)
			}
		}
		utxo := makeVSWithValueFee(t, ownerSigner, 1, valueOut, vsfee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}

	// Setup Tx
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed objects
	for k := 0; k < totalInputsAndOutputs; k++ {
		err = consumedUTXOs[k].valueStore.Sign(vin[k], ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Tx with 16->16 value transfer")
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxValueTransferSize3(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make inputs and outputs
	totalInputsAndOutputs := constants.MaxTxVectorLength
	consumedUTXOs := Vout{}
	vin := Vin{}
	vout := Vout{}
	value64 := uint64(constants.MaxUint64)
	for k := 0; k < totalInputsAndOutputs; k++ {
		v := value64 - uint64(k)
		valueIn, err := new(uint256.Uint256).FromUint64(v)
		if err != nil {
			t.Fatal(err)
		}
		// Make consumed utxo
		consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
		valueBytes := utils.MarshalUint64(v)
		indexBytes := utils.MarshalUint64(uint64(k))
		vHash := crypto.Hasher(valueBytes, indexBytes)
		err = consumedUTXO.SetTxHash(vHash)
		if err != nil {
			t.Fatal(err)
		}
		consumedUTXOs = append(consumedUTXOs, consumedUTXO)
		txin, err := consumedUTXO.MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		vin = append(vin, txin)
		// Make corresponding utxo
		valueOut, err := new(uint256.Uint256).Sub(valueIn, vsfee)
		if err != nil {
			t.Fatal(err)
		}
		if k == 0 {
			// Subtract txfee from initial
			_, err = valueOut.Sub(valueOut, minTxFee)
			if err != nil {
				t.Fatal(err)
			}
		}
		utxo := makeVSWithValueFee(t, ownerSigner, 1, valueOut, vsfee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}

	// Setup Tx
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign objects
	for k := 0; k < totalInputsAndOutputs; k++ {
		err = consumedUTXOs[k].valueStore.Sign(vin[k], ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Tx with max->max value transfer")
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxCleanupTxSize1(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make inputs
	totalCleanup := 1
	consumedUTXOs := Vout{}
	vin := Vin{}
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	for k := 0; k < totalCleanup; k++ {
		// Need to make datastores to clean up
		rawData := crypto.Hasher([]byte("RawData"), utils.MarshalUint64(uint64(k)))
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		consumedUTXO := makeDSWithValueFee(t, ownerSigner, 1, rawData, index, startEpoch, numEpochs, dsfee)
		indexBytes := utils.MarshalUint64(uint64(k))
		vHash := crypto.Hasher([]byte("TxHashBytes"), indexBytes)
		err = consumedUTXO.SetTxHash(vHash)
		if err != nil {
			t.Fatal(err)
		}
		consumedUTXOs = append(consumedUTXOs, consumedUTXO)
		txin, err := consumedUTXO.MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		vin = append(vin, txin)
	}

	// Make outputs
	// Compute remainingValue to have correct ValueStore
	cleanupFee := uint256.Zero()
	currentHeight := constants.EpochLength*(startEpoch+numEpochs) + 1
	remainingValue, err := consumedUTXOs.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	utxo := makeVSWithValueFee(t, ownerSigner, 1, remainingValue, cleanupFee)

	vout := []*TXOut{utxo}

	// Setup Tx
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign objects
	for k := 0; k < totalCleanup; k++ {
		err = consumedUTXOs[k].dataStore.Sign(vin[k], ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}
	if !tx.IsCleanupTx(currentHeight, consumedUTXOs) {
		t.Fatal("Should be a cleanup tx!")
	}

	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CleanupTx with 1 datastore")
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxCleanupTxSize2(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make inputs
	totalCleanup := 16
	consumedUTXOs := Vout{}
	vin := Vin{}
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	for k := 0; k < totalCleanup; k++ {
		// Need to make datastores to clean up
		rawData := crypto.Hasher([]byte("RawData"), utils.MarshalUint64(uint64(k)))
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		consumedUTXO := makeDSWithValueFee(t, ownerSigner, 1, rawData, index, startEpoch, numEpochs, dsfee)
		indexBytes := utils.MarshalUint64(uint64(k))
		vHash := crypto.Hasher([]byte("TxHashBytes"), indexBytes)
		err = consumedUTXO.SetTxHash(vHash)
		if err != nil {
			t.Fatal(err)
		}
		consumedUTXOs = append(consumedUTXOs, consumedUTXO)
		txin, err := consumedUTXO.MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		vin = append(vin, txin)
	}

	// Make outputs
	// Compute remainingValue to have correct ValueStore
	cleanupFee := uint256.Zero()
	currentHeight := constants.EpochLength*(startEpoch+numEpochs) + 1
	remainingValue, err := consumedUTXOs.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	utxo := makeVSWithValueFee(t, ownerSigner, 1, remainingValue, cleanupFee)

	vout := []*TXOut{utxo}

	// Setup Tx
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign objects
	for k := 0; k < totalCleanup; k++ {
		err = consumedUTXOs[k].dataStore.Sign(vin[k], ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}
	if !tx.IsCleanupTx(currentHeight, consumedUTXOs) {
		t.Fatal("Should be a cleanup tx!")
	}

	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CleanupTx with 16 datastores")
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxCleanupTxSize3(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make inputs
	totalCleanup := constants.MaxTxVectorLength
	consumedUTXOs := Vout{}
	vin := Vin{}
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	for k := 0; k < totalCleanup; k++ {
		// Need to make datastores to clean up
		rawData := crypto.Hasher([]byte("RawData"), utils.MarshalUint64(uint64(k)))
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		consumedUTXO := makeDSWithValueFee(t, ownerSigner, 1, rawData, index, startEpoch, numEpochs, dsfee)
		indexBytes := utils.MarshalUint64(uint64(k))
		vHash := crypto.Hasher([]byte("TxHashBytes"), indexBytes)
		err = consumedUTXO.SetTxHash(vHash)
		if err != nil {
			t.Fatal(err)
		}
		consumedUTXOs = append(consumedUTXOs, consumedUTXO)
		txin, err := consumedUTXO.MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		vin = append(vin, txin)
	}

	// Make outputs
	// Compute remainingValue to have correct ValueStore
	cleanupFee := uint256.Zero()
	currentHeight := constants.EpochLength*(startEpoch+numEpochs) + 1
	remainingValue, err := consumedUTXOs.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	utxo := makeVSWithValueFee(t, ownerSigner, 1, remainingValue, cleanupFee)

	vout := []*TXOut{utxo}

	// Setup Tx
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign objects
	for k := 0; k < totalCleanup; k++ {
		err = consumedUTXOs[k].dataStore.Sign(vin[k], ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}
	if !tx.IsCleanupTx(currentHeight, consumedUTXOs) {
		t.Fatal("Should be a cleanup tx!")
	}

	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("CleanupTx with max datastores")
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize1(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 128
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(0)))
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, startEpoch, numEpochs, dsfee)
	dsValuePlusFee, err := utxo1.ValuePlusFee()
	if err != nil {
		t.Fatal(err)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, dsValuePlusFee)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)

	// Setup Tx
	vin := []*TXIn{txin}
	vout := []*TXOut{utxo1, utxo2}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	err = utxo1.dataStore.PreSign(ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->1+1; storing %v bytes\n", dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize2(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 1024
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(0)))
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, startEpoch, numEpochs, dsfee)
	dsValuePlusFee, err := utxo1.ValuePlusFee()
	if err != nil {
		t.Fatal(err)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, dsValuePlusFee)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)

	// Setup Tx
	vin := []*TXIn{txin}
	vout := []*TXOut{utxo1, utxo2}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	err = utxo1.dataStore.PreSign(ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->1+1; storing %v bytes\n", dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize3(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := int(constants.MaxDataStoreSize)
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(0)))
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, startEpoch, numEpochs, dsfee)
	dsValuePlusFee, err := utxo1.ValuePlusFee()
	if err != nil {
		t.Fatal(err)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, dsValuePlusFee)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)

	// Setup Tx
	vin := []*TXIn{txin}
	vout := []*TXOut{utxo1, utxo2}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	err = utxo1.dataStore.PreSign(ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->1+1; storing %v bytes\n", dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize4(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 128
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	numOutputs := 1
	if numOutputs > constants.MaxTxVectorLength-1 {
		t.Fatal("Too many outputs!")
	}
	vout := Vout{}
	totalOutput := uint256.Zero()
	for k := 0; k < numOutputs; k++ {
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		utxo := makeDSWithValueFee(t, ownerSigner, 0, utils.CopySlice(rawData), index, startEpoch, numEpochs, dsfee)
		dsValuePlusFee, err := utxo.ValuePlusFee()
		if err != nil {
			t.Fatal(err)
		}
		_, err = totalOutput.Add(totalOutput, dsValuePlusFee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, totalOutput)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)
	vout = append(vout, utxo2)

	// Setup Tx
	vin := []*TXIn{txin}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	for k := 0; k < numOutputs; k++ {
		err = vout[k].dataStore.PreSign(ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->%v+1; storing %v bytes\n", numOutputs, dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize5(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 128
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	numOutputs := constants.MaxTxVectorLength - 1
	if numOutputs > constants.MaxTxVectorLength-1 {
		t.Fatal("Too many outputs!")
	}
	vout := Vout{}
	totalOutput := uint256.Zero()
	for k := 0; k < numOutputs; k++ {
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		utxo := makeDSWithValueFee(t, ownerSigner, 0, utils.CopySlice(rawData), index, startEpoch, numEpochs, dsfee)
		dsValuePlusFee, err := utxo.ValuePlusFee()
		if err != nil {
			t.Fatal(err)
		}
		_, err = totalOutput.Add(totalOutput, dsValuePlusFee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, totalOutput)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)
	vout = append(vout, utxo2)

	// Setup Tx
	vin := []*TXIn{txin}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	for k := 0; k < numOutputs; k++ {
		err = vout[k].dataStore.PreSign(ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->%v+1; storing %v bytes\n", numOutputs, dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize6(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 32
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	numOutputs := constants.MaxTxVectorLength - 1
	if numOutputs > constants.MaxTxVectorLength-1 {
		t.Fatal("Too many outputs!")
	}
	vout := Vout{}
	totalOutput := uint256.Zero()
	for k := 0; k < numOutputs; k++ {
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		utxo := makeDSWithValueFee(t, ownerSigner, 0, utils.CopySlice(rawData), index, startEpoch, numEpochs, dsfee)
		dsValuePlusFee, err := utxo.ValuePlusFee()
		if err != nil {
			t.Fatal(err)
		}
		_, err = totalOutput.Add(totalOutput, dsValuePlusFee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, totalOutput)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)
	vout = append(vout, utxo2)

	// Setup Tx
	vin := []*TXIn{txin}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	for k := 0; k < numOutputs; k++ {
		err = vout[k].dataStore.PreSign(ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->%v+1; storing %v bytes\n", numOutputs, dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize7(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 32
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	numOutputs := 1
	if numOutputs > constants.MaxTxVectorLength-1 {
		t.Fatal("Too many outputs!")
	}
	vout := Vout{}
	totalOutput := uint256.Zero()
	for k := 0; k < numOutputs; k++ {
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		utxo := makeDSWithValueFee(t, ownerSigner, 0, utils.CopySlice(rawData), index, startEpoch, numEpochs, dsfee)
		dsValuePlusFee, err := utxo.ValuePlusFee()
		if err != nil {
			t.Fatal(err)
		}
		_, err = totalOutput.Add(totalOutput, dsValuePlusFee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, totalOutput)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)
	vout = append(vout, utxo2)

	// Setup Tx
	vin := []*TXIn{txin}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	for k := 0; k < numOutputs; k++ {
		err = vout[k].dataStore.PreSign(ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->%v+1; storing %v bytes\n", numOutputs, dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize8(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 256
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	numOutputs := 1
	if numOutputs > constants.MaxTxVectorLength-1 {
		t.Fatal("Too many outputs!")
	}
	vout := Vout{}
	totalOutput := uint256.Zero()
	for k := 0; k < numOutputs; k++ {
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		utxo := makeDSWithValueFee(t, ownerSigner, 0, utils.CopySlice(rawData), index, startEpoch, numEpochs, dsfee)
		dsValuePlusFee, err := utxo.ValuePlusFee()
		if err != nil {
			t.Fatal(err)
		}
		_, err = totalOutput.Add(totalOutput, dsValuePlusFee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, totalOutput)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)
	vout = append(vout, utxo2)

	// Setup Tx
	vin := []*TXIn{txin}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	for k := 0; k < numOutputs; k++ {
		err = vout[k].dataStore.PreSign(ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->%v+1; storing %v bytes\n", numOutputs, dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize9(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 256
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	numOutputs := constants.MaxTxVectorLength - 1
	if numOutputs > constants.MaxTxVectorLength-1 {
		t.Fatal("Too many outputs!")
	}
	vout := Vout{}
	totalOutput := uint256.Zero()
	for k := 0; k < numOutputs; k++ {
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		utxo := makeDSWithValueFee(t, ownerSigner, 0, utils.CopySlice(rawData), index, startEpoch, numEpochs, dsfee)
		dsValuePlusFee, err := utxo.ValuePlusFee()
		if err != nil {
			t.Fatal(err)
		}
		_, err = totalOutput.Add(totalOutput, dsValuePlusFee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, totalOutput)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)
	vout = append(vout, utxo2)

	// Setup Tx
	vin := []*TXIn{txin}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	for k := 0; k < numOutputs; k++ {
		err = vout[k].dataStore.PreSign(ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->%v+1; storing %v bytes\n", numOutputs, dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize10(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 256
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	numOutputs := 15
	if numOutputs > constants.MaxTxVectorLength-1 {
		t.Fatal("Too many outputs!")
	}
	vout := Vout{}
	totalOutput := uint256.Zero()
	for k := 0; k < numOutputs; k++ {
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		utxo := makeDSWithValueFee(t, ownerSigner, 0, utils.CopySlice(rawData), index, startEpoch, numEpochs, dsfee)
		dsValuePlusFee, err := utxo.ValuePlusFee()
		if err != nil {
			t.Fatal(err)
		}
		_, err = totalOutput.Add(totalOutput, dsValuePlusFee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, totalOutput)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)
	vout = append(vout, utxo2)

	// Setup Tx
	vin := []*TXIn{txin}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	for k := 0; k < numOutputs; k++ {
		err = vout[k].dataStore.PreSign(ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->%v+1; storing %v bytes\n", numOutputs, dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize11(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 128
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	numOutputs := 15
	if numOutputs > constants.MaxTxVectorLength-1 {
		t.Fatal("Too many outputs!")
	}
	vout := Vout{}
	totalOutput := uint256.Zero()
	for k := 0; k < numOutputs; k++ {
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		utxo := makeDSWithValueFee(t, ownerSigner, 0, utils.CopySlice(rawData), index, startEpoch, numEpochs, dsfee)
		dsValuePlusFee, err := utxo.ValuePlusFee()
		if err != nil {
			t.Fatal(err)
		}
		_, err = totalOutput.Add(totalOutput, dsValuePlusFee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, totalOutput)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)
	vout = append(vout, utxo2)

	// Setup Tx
	vin := []*TXIn{txin}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	for k := 0; k < numOutputs; k++ {
		err = vout[k].dataStore.PreSign(ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->%v+1; storing %v bytes\n", numOutputs, dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}

func TestTxDataStoreSize12(t *testing.T) {
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(3)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(4)
	msg.SetMinTxFeeCostRatio(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	// Setup fees
	vsfee, err := new(uint256.Uint256).FromBigInt(vsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	dsfee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	minTxFee, err := new(uint256.Uint256).FromBigInt(tfFeeBig)
	if err != nil {
		t.Fatal(err)
	}

	// Make input
	value64 := uint64(constants.MaxUint64)
	valueIn, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	consumedUTXO := makeVSWithValueFee(t, ownerSigner, 1, valueIn, vsfee)
	txin, err := consumedUTXO.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Sum value of inputs
	totalInput := uint256.Zero()
	_, err = totalInput.Add(totalInput, valueIn)
	if err != nil {
		t.Fatal(err)
	}

	// Make output
	startEpoch := uint32(1)
	numEpochs := uint32(1)
	dataSize := 32
	rawData := make([]byte, dataSize)
	for k := 0; k < dataSize; k++ {
		// Fill with 1s
		rawData[k] = 255
	}
	numOutputs := 15
	if numOutputs > constants.MaxTxVectorLength-1 {
		t.Fatal("Too many outputs!")
	}
	vout := Vout{}
	totalOutput := uint256.Zero()
	for k := 0; k < numOutputs; k++ {
		index := crypto.Hasher([]byte("Index"), utils.MarshalUint64(uint64(k)))
		utxo := makeDSWithValueFee(t, ownerSigner, 0, utils.CopySlice(rawData), index, startEpoch, numEpochs, dsfee)
		dsValuePlusFee, err := utxo.ValuePlusFee()
		if err != nil {
			t.Fatal(err)
		}
		_, err = totalOutput.Add(totalOutput, dsValuePlusFee)
		if err != nil {
			t.Fatal(err)
		}
		vout = append(vout, utxo)
	}
	valueOut := totalInput.Clone()
	_, err = valueOut.Sub(valueOut, minTxFee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = valueOut.Sub(valueOut, totalOutput)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 0, valueOut, vsfee)
	vout = append(vout, utxo2)

	// Setup Tx
	vin := []*TXIn{txin}
	refUTXOs := Vout{consumedUTXO}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  minTxFee.Clone(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	// Sign consumed object object
	err = consumedUTXO.valueStore.Sign(txin, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	// Sign output DataStore
	for k := 0; k < numOutputs; k++ {
		err = vout[k].dataStore.PreSign(ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	currentHeight := uint32(1)
	chainID := uint32(2)
	err = tx.PreValidatePending(chainID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Validate(nil, currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.PostValidatePending(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
	txBytes, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("Tx DataStore with 1->%v+1; storing %v bytes\n", numOutputs, dataSize)
	fmt.Println(len(txBytes))
	fmt.Println()
}
