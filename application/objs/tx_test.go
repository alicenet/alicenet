package objs

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"strconv"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/stretchr/testify/assert"
)

func makeVS(t *testing.T, ownerSigner Signer, i int) *TXOut {
	cid := uint32(2)
	val := uint256.One()

	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	fee := new(uint256.Uint256)
	vsp := &VSPreImage{
		ChainID: cid,
		Value:   val,
		Owner:   owner,
		Fee:     fee.Clone(),
	}
	var txHash []byte
	if i == 0 {
		txHash = make([]byte, constants.HashLen)
	} else {
		txHash = crypto.Hasher([]byte(strconv.Itoa(i)))
	}
	vs := &ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	vsEqual(t, vs, vs2)
	utxInputs := &TXOut{}
	err = utxInputs.NewValueStore(vs)
	if err != nil {
		t.Fatal(err)
	}
	return utxInputs
}

func makeVSWithValueFee(t *testing.T, ownerSigner Signer, i int, value, fee *uint256.Uint256) *TXOut {
	if value == nil || fee == nil {
		panic("invalid value or fee")
	}
	cid := uint32(2)

	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &ValueStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	vsp := &VSPreImage{
		ChainID: cid,
		Value:   value,
		Owner:   owner,
		Fee:     fee.Clone(),
	}
	var txHash []byte
	if i == 0 {
		txHash = make([]byte, constants.HashLen)
	} else {
		txHash = crypto.Hasher([]byte(strconv.Itoa(i)))
	}
	vs := &ValueStore{
		VSPreImage: vsp,
		TxHash:     txHash,
	}
	vs2 := &ValueStore{}
	vsBytes, err := vs.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = vs2.UnmarshalBinary(vsBytes)
	if err != nil {
		t.Fatal(err)
	}
	vsEqual(t, vs, vs2)
	utxInputs := &TXOut{}
	err = utxInputs.NewValueStore(vs)
	if err != nil {
		t.Fatal(err)
	}
	return utxInputs
}

func makeDSWithValueFee(t *testing.T, ownerSigner Signer, i int, rawData []byte, index []byte, startEpoch uint32, numEpochs uint32, fee *uint256.Uint256) *TXOut {
	if fee == nil || len(rawData) == 0 {
		panic("invalid fee or rawData")
	}
	cid := uint32(2)

	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &DataStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	dataSize32 := uint32(len(rawData))
	deposit, err := BaseDepositEquation(dataSize32, numEpochs)
	if err != nil {
		t.Fatal(err)
	}

	dsp := &DSPreImage{
		ChainID:  cid,
		Index:    index,
		IssuedAt: startEpoch,
		Deposit:  deposit.Clone(),
		RawData:  rawData,
		Owner:    owner,
		Fee:      fee.Clone(),
	}
	err = dsp.ValidateDeposit()
	if err != nil {
		t.Fatal(err)
	}
	var txHash []byte
	if i == 0 {
		txHash = make([]byte, constants.HashLen)
	} else {
		txHash = crypto.Hasher([]byte(strconv.Itoa(i)))
	}
	dsl := &DSLinker{
		DSPreImage: dsp,
		TxHash:     txHash,
	}
	ds := &DataStore{
		DSLinker: dsl,
	}
	err = ds.PreSign(ownerSigner)
	if err != nil {
		t.Fatal(err)
	}
	err = ds.ValidatePreSignature()
	if err != nil {
		t.Fatal(err)
	}

	ds2 := &DataStore{}
	dsBytes, err := ds.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = ds2.UnmarshalBinary(dsBytes)
	if err != nil {
		t.Fatal(err)
	}
	dsEqual(t, ds, ds2)
	utxInputs := &TXOut{}
	err = utxInputs.NewDataStore(ds)
	if err != nil {
		t.Fatal(err)
	}
	return utxInputs
}

func TestTx(t *testing.T) {
	msg := MakeMockStorageGetter()
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	consumedUTXOs := Vout{}
	for i := 1; i < 5; i++ {
		consumedUTXOs = append(consumedUTXOs, makeVS(t, ownerSigner, i))
	}
	err := consumedUTXOs.SetTxOutIdx()
	if err != nil {
		t.Fatal(err)
	}

	txInputs := []*TXIn{}
	for i := 0; i < 4; i++ {
		txIn, err := consumedUTXOs[i].MakeTxIn()
		if err != nil {
			t.Fatal(err)
		}
		txInputs = append(txInputs, txIn)
	}
	generatedUTXOs := Vout{}
	for i := 1; i < 5; i++ {
		generatedUTXOs = append(generatedUTXOs, makeVS(t, ownerSigner, 0))
	}
	err = generatedUTXOs.SetTxOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	tx := &Tx{
		Vin:  txInputs,
		Vout: generatedUTXOs,
		Fee:  uint256.Zero(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 4; i++ {
		err = consumedUTXOs[i].valueStore.Sign(tx.Vin[i], ownerSigner)
		if err != nil {
			t.Fatal(err)
		}
	}

	txb, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("TX in hex: %x", txb)
	for _, utxo := range tx.Vout {
		vs, _ := utxo.ValueStore()
		uid, err := utxo.UTXOID()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("ValueStore: ChainID: %v\n", vs.VSPreImage.ChainID)
		vsValue, err := vs.Value()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("ValueStore: Next: %v\n", vsValue)
		t.Logf("ValueStore: TxHash: %x\n", vs.TxHash)
		t.Logf("ValueStore: Owner: %x\n", vs.VSPreImage.Owner.Account)
		t.Logf("ValueStore: UTXOID: %x\n", uid)
		sig, err := vs.VSPreImage.Owner.MarshalBinary()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("ValueStore: Owner: %x\n", sig)
	}
	tx2 := &Tx{}
	err = tx2.UnmarshalBinary(txb)
	if err != nil {
		t.Fatal(err)
	}

	// check marshaling did not change data
	txh, err := tx.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	txh2, err := tx2.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(txh2, txh) {
		t.Fatal()
	}

	// validate the returned object
	_, err = tx2.ValidateUnique(nil)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.ValidateEqualVinVout(1, consumedUTXOs)
	if err != nil {
		t.Fatal(err)
	}
	err = tx.ValidateTxHash()
	if err != nil {
		t.Fatal(err)
	}
	err = tx2.ValidatePreSignature()
	if err != nil {
		t.Fatal(err)
	}
	err = tx2.ValidateSignature(1, consumedUTXOs)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx2.Validate(nil, 1, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}

	txVec := TxVec([]*Tx{tx})
	err = txVec.Validate(1, consumedUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}

	_, err = txVec.ConsumedUTXOID()
	if err != nil {
		t.Fatal(err)
	}

	_, err = txVec.GeneratedUTXOID()
	if err != nil {
		t.Fatal(err)
	}

	_, err = txVec.GeneratedPreHash()
	if err != nil {
		t.Fatal(err)
	}

	isDep := txVec.ConsumedIsDeposit()
	for _, i := range isDep {
		if i {
			t.Fatalf("%v", i)
		}
	}

	// check indexing
	txVec = append(txVec, []*Tx{tx}...)
	_, err = txVec.ConsumedUTXOID()
	if err != nil {
		t.Fatal(err)
	}

	_, err = txVec.GeneratedUTXOID()
	if err != nil {
		t.Fatal(err)
	}

	_, err = txVec.GeneratedPreHash()
	if err != nil {
		t.Fatal(err)
	}

	isDep = txVec.ConsumedIsDeposit()
	for _, i := range isDep {
		if i {
			t.Fatalf("%v", i)
		}
	}

	privk, err := hex.DecodeString("2da4ef21b864d2cc526dbdb2a120bd2874c36c9d0a1fb7f8c63d7f7a8b41de8f")
	if err != nil {
		t.Fatal(err)
	}

	signer := &crypto.BNSigner{}
	err = signer.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	pubk, _ := signer.Pubkey()
	account := crypto.GetAccount(pubk)

	txin := &TXIn{
		TXInLinker: &TXInLinker{
			TXInPreImage: &TXInPreImage{
				ChainID:        2,
				ConsumedTxIdx:  0,
				ConsumedTxHash: crypto.Hasher(make([]byte, constants.HashLen)),
			},
		},
	}

	txib, err := txin.TXInLinker.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	fee := new(uint256.Uint256)
	val, err := new(uint256.Uint256).FromUint64(300)
	if err != nil {
		t.Fatal(err)
	}
	vs := &ValueStore{
		VSPreImage: &VSPreImage{
			Value:    val, // 300
			TXOutIdx: 0,
			ChainID:  1,
			Owner: &ValueStoreOwner{
				SVA:       ValueStoreSVA,
				CurveSpec: constants.CurveBN256Eth,
				Account:   account,
			},
			Fee: fee.Clone(),
		},
	}

	utxo := &TXOut{}
	err = utxo.NewValueStore(vs)
	if err != nil {
		t.Fatal(err)
	}

	tx = &Tx{
		Vin:  Vin{txin},
		Vout: Vout{utxo},
		Fee:  uint256.Zero(),
	}

	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}

	sig, err := signer.Sign(txib)
	if err != nil {
		t.Fatal(err)
	}

	s1 := &ValueStoreSignature{
		SVA:       ValueStoreSVA,
		CurveSpec: constants.CurveBN256Eth,
		Signature: sig,
	}
	s1b, err := s1.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	txin.Signature = s1b

	rawb, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%x", rawb)

	tx33 := &Tx{}
	tx33b, _ := hex.DecodeString("00000000000002000500000017000000a1000000170000000400000000000200040000000000020031000000120600000400000001000100190000000201000002000000000000000100000002010000290decd9548b62a8d60345a988386fc84ba6bc95484008f6362f93160ef3e5639e392803cea0aeba760ab78d2f3dc0231bbf2fea9a5a1d3e62f230849f1ba8a20102230120659dd94202a064848b83886699a6388d894495cd1e9e200f2ef261a2d72f9ea359d2684ea5f7a5ae6b46ed2ebcd64f517255e8f1a3b4872b8118129712084c325f093bfe6b9341b102f5bf07cf21effcb50104351c594f94927dfcf6f92ba4a43557597fc7d21a74b7a8874dc787bb5b25c764b2a0b52969be4901f85e085db78b91f901046842c899820834e2df91ae4f9169544715e48d0df2fad405134ad114f827e45cda472177690a30395dd4262ba525925cac420f956221c1de00000000000004000000010001000100000000000000000000000000020004000000010001001500000002010000010000002c01000001000000b200000001028e80cf09fc395986a2e9a73b84e00018e64131b100009e392803cea0aeba760ab78d2f3dc0231bbf2fea9a5a1d3e62f230849f1ba8a2")
	err = tx33.UnmarshalBinary(tx33b)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("ConsumedTxHash: %x", tx.Vin[0].TXInLinker.TXInPreImage.ConsumedTxHash)
	t.Logf("ConsumedTxIdx: %x", tx.Vin[0].TXInLinker.TXInPreImage.ConsumedTxIdx)
	t.Logf("ChainID: %x", tx.Vin[0].TXInLinker.TXInPreImage.ChainID)
	t.Logf("TxHash: %x", tx.Vin[0].TXInLinker.TxHash)
	t.Logf("Sig: %x", tx.Vin[0].Signature)
}

func TestTxMarshalGood1(t *testing.T) {
	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo1.valueStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}

	txb, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	tx2 := &Tx{}
	err = tx2.UnmarshalBinary(txb)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxMarshalGood2(t *testing.T) {
	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(1)
	vsfee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, vsfee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo1.valueStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	txfee := uint256.Two()
	newValue, err := new(uint256.Uint256).Sub(value, vsfee)
	if err != nil {
		t.Fatal(err)
	}
	_, err = newValue.Sub(newValue, txfee)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, vsfee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = txfee

	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}

	txb, err := tx.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	tx2 := &Tx{}
	err = tx2.UnmarshalBinary(txb)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxMarshalBad1(t *testing.T) {
	tx := &Tx{}
	_, err := tx.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxMarshalBad2(t *testing.T) {
	txin := &TXIn{}
	vin := Vin{}
	vout := Vout{}
	for i := 0; i < constants.MaxTxVectorLength+1; i++ {
		vin = append(vin, txin)
	}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	_, err := tx.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	utxo := &TXOut{}
	vin = Vin{}
	vout = Vout{}
	for i := 0; i < constants.MaxTxVectorLength+1; i++ {
		vout = append(vout, utxo)
	}
	tx = &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	_, err = tx.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestTxValidateChainIDBad1(t *testing.T) {
	chainID := uint32(1)
	tx := &Tx{}
	err := tx.ValidateChainID(chainID)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateChainIDBad2(t *testing.T) {
	chainID := uint32(0)
	tx := &Tx{
		Vin:  Vin{&TXIn{}},
		Vout: Vout{&TXOut{}},
		Fee:  uint256.Zero(),
	}
	err := tx.ValidateChainID(chainID)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxCannotBeMinedUntilBad(t *testing.T) {
	tx := &Tx{}
	_, err := tx.CannotBeMinedUntil()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateIssuedAtForMiningGood(t *testing.T) {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo1.valueStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	tx := &Tx{}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	currentHeight := uint32(1)
	err = tx.ValidateIssuedAtForMining(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxEpochOfExpirationForMiningBad(t *testing.T) {
	tx := &Tx{}
	_, err := tx.EpochOfExpirationForMining()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxEpochOfExpirationForMiningGood(t *testing.T) {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo1.valueStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	tx := &Tx{}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	height, err := tx.EpochOfExpirationForMining()
	if err != nil {
		t.Fatal(err)
	}
	if height != constants.MaxUint32 {
		t.Fatal("Incorrect height")
	}
}

func TestTxValidateIssuedAtForMiningBad(t *testing.T) {
	currentHeight := uint32(1)
	tx := &Tx{}
	err := tx.ValidateIssuedAtForMining(currentHeight)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateUniqueGood(t *testing.T) {
	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo1.valueStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	_, err = tx.ValidateUnique(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxValidateUniqueBad1(t *testing.T) {
	tx := &Tx{}

	txin1 := &TXIn{}
	tx.Vin = []*TXIn{txin1}

	_, err := tx.ValidateUnique(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateUniqueBad2(t *testing.T) {
	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo1.valueStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := &TXOut{}
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	_, err = tx.ValidateUnique(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateUniqueBad3(t *testing.T) {
	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo1.valueStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1, txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	_, err = tx.ValidateUnique(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateUniqueBad4(t *testing.T) {
	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo1.valueStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2, utxo2}
	tx.Fee = uint256.Zero()

	_, err = tx.ValidateUnique(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateDataStoreIndexesGood1(t *testing.T) {
	tx := &Tx{}
	_, err := tx.ValidateDataStoreIndexes(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxValidateDataStoreIndexesGood2(t *testing.T) {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	index := make([]byte, constants.HashLen)
	index[0] = 1

	fee := uint256.Zero()
	rawData := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, iat, numEpochs, fee)

	tx := &Tx{
		Vout: Vout{utxo1},
	}
	_, err := tx.ValidateDataStoreIndexes(nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxValidateDataStoreIndexesBad1(t *testing.T) {
	utxo := &TXOut{}
	utxo.hasDataStore = true
	tx := &Tx{
		Vout: Vout{utxo},
	}
	_, err := tx.ValidateDataStoreIndexes(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateDataStoreIndexesBad2(t *testing.T) {
	ds := &DataStore{}
	utxo := &TXOut{}
	err := utxo.NewDataStore(ds)
	assert.Nil(t, err)
	tx := &Tx{
		Vout: Vout{utxo},
	}
	_, err = tx.ValidateDataStoreIndexes(nil)
	assert.NotNil(t, err)
}

func TestTxValidateDataStoreIndexesBad3(t *testing.T) {
	ds := &DataStore{}
	ds.DSLinker = &DSLinker{}
	ds.DSLinker.DSPreImage = &DSPreImage{}
	ds.DSLinker.DSPreImage.Index = make([]byte, constants.HashLen)
	utxo := &TXOut{}
	err := utxo.NewDataStore(ds)
	assert.Nil(t, err)
	tx := &Tx{
		Vout: Vout{utxo},
	}
	_, err = tx.ValidateDataStoreIndexes(nil)
	assert.NotNil(t, err)
}

func TestTxValidateDataStoreIndexesBad4(t *testing.T) {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	index := make([]byte, constants.HashLen)
	index[0] = 1

	fee := uint256.Zero()
	rawData := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, iat, numEpochs, fee)

	tx := &Tx{
		Vout: Vout{utxo1, utxo1},
	}
	_, err := tx.ValidateDataStoreIndexes(nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxConsumedPreHash(t *testing.T) {
	tx := &Tx{}
	_, err := tx.ConsumedPreHash()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTxConsumedUTXOID(t *testing.T) {
	tx := &Tx{}
	_, err := tx.ConsumedUTXOID()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTxGeneratedUTXOID(t *testing.T) {
	tx := &Tx{}
	_, err := tx.GeneratedUTXOID()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTxGeneratedPreHash(t *testing.T) {
	tx := &Tx{}
	_, err := tx.GeneratedPreHash()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTxValidateSignature(t *testing.T) {
	currentHeight := uint32(1)
	refUTXOs := Vout{}
	tx := &Tx{}
	err := tx.ValidateSignature(currentHeight, refUTXOs)
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTxValidatePreSignature(t *testing.T) {
	tx := &Tx{}
	err := tx.ValidatePreSignature()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}

func TestTxCallTxHash(t *testing.T) {
	tx := &Tx{}
	tx.Fee = uint256.Zero()
	hashTrue := crypto.Hasher([][]byte{}...)
	hash, err := tx.TxHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(hash, hashTrue) {
		t.Fatal("txhashes do not match")
	}
}

func TestTxCallTxHashBad1(t *testing.T) {
	txin := &TXIn{}
	tx := &Tx{
		Vin: Vin{txin},
	}
	_, err := tx.TxHash()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxCallTxHashBad2(t *testing.T) {
	utxo := &TXOut{}
	tx := &Tx{
		Vout: Vout{utxo},
	}
	_, err := tx.TxHash()
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateFeesGood1(t *testing.T) {
	msg := MakeMockStorageGetter()
	storage := MakeStorage(msg)

	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	err = tx.ValidateFees(0, nil, storage)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxValidateFeesGood2(t *testing.T) {
	// Is valid CleanupTx; Validate the fees
	msg := MakeMockStorageGetter()
	msg.SetMinTxFee(big.NewInt(1))
	storage := MakeStorage(msg)
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	fee := uint256.Zero()
	index := make([]byte, constants.HashLen)
	index[0] = 1
	rawData := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, iat, numEpochs, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Compute remainingValue to have correct ValueStore
	currentHeight := constants.EpochLength*(iat+numEpochs) + 1
	remainingValue, err := utxo1.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 1, remainingValue, fee)

	vin := []*TXIn{txin1}
	refUTXOs := []*TXOut{utxo1}
	vout := []*TXOut{utxo2}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if !cleanup {
		t.Fatal("Should be valid CleanupTx")
	}
	err = tx.ValidateFees(currentHeight, refUTXOs, storage)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTxValidateFeesBad1(t *testing.T) {
	msg := MakeMockStorageGetter()
	storage := MakeStorage(msg)

	tx := &Tx{}
	err := tx.ValidateFees(0, nil, storage)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateFeesBad2(t *testing.T) {
	txin := &TXIn{}
	utxo := &TXOut{}
	tx := &Tx{
		Vin:  []*TXIn{txin},
		Vout: []*TXOut{utxo},
		Fee:  uint256.Zero(),
	}
	err := tx.ValidateFees(0, nil, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

// TODO: look at this more!
func TestTxValidateFeesBad3(t *testing.T) {
	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	msg := MakeMockStorageGetter()
	minTxFee := big.NewInt(1)
	msg.SetMinTxFee(minTxFee)
	storage := MakeStorage(msg)
	err = tx.ValidateFees(0, Vout{utxo1}, storage)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateFeesBad4(t *testing.T) {
	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	msg := MakeMockStorageGetter()
	minTxFee := big.NewInt(1)
	msg.SetMinTxFee(minTxFee)
	storage := MakeStorage(msg)
	err = tx.ValidateFees(0, nil, storage)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateFeesBad5(t *testing.T) {
	// Raise an error for invalid storage call
	tx := &Tx{}

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := makeVSWithValueFee(t, ownerSigner, 2, value, fee)
	tx.Vin = []*TXIn{txin1}
	tx.Vout = []*TXOut{utxo2}
	tx.Fee = uint256.Zero()

	err = tx.ValidateFees(0, nil, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxIsCleanupTxBad1(t *testing.T) {
	// Invalid Tx; fails because no Fee
	tx := &Tx{}
	currentHeight := uint32(0)
	refUTXOs := Vout{}
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if cleanup {
		t.Fatal("Should not be CleanupTx")
	}
}

func TestTxIsCleanupTxBad2(t *testing.T) {
	// Invalid Vin; no DataStore
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	value64 := uint64(10000)
	value, err := new(uint256.Uint256).FromUint64(value64)
	if err != nil {
		t.Fatal(err)
	}
	fee64 := uint64(0)
	fee, err := new(uint256.Uint256).FromUint64(fee64)
	if err != nil {
		t.Fatal(err)
	}
	utxo1 := makeVSWithValueFee(t, ownerSigner, 1, value, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Invalid Vin; not DataStore
	vin := []*TXIn{txin1}
	refUTXOs := []*TXOut{utxo1}
	tx := &Tx{
		Vin:  vin,
		Vout: Vout{},
		Fee:  uint256.Zero(),
	}
	currentHeight := uint32(1)
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if cleanup {
		t.Fatal("Should not be CleanupTx")
	}
}

func TestTxIsCleanupTxBad3(t *testing.T) {
	// Must have valid Vin and invalid Vout (not present)
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	index := make([]byte, constants.HashLen)
	index[0] = 1

	fee := uint256.Zero()
	rawData := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, iat, numEpochs, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	vin := []*TXIn{txin1}
	refUTXOs := []*TXOut{utxo1}
	tx := &Tx{
		Vin:  vin,
		Vout: Vout{},
		Fee:  uint256.Zero(),
	}
	currentHeight := constants.EpochLength*(iat+numEpochs) + 1
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if cleanup {
		t.Fatal("Should not be CleanupTx")
	}
}

func TestTxIsCleanupTxBad4(t *testing.T) {
	// Must have valid Vin and invalid Vout (incorrect utxo type)
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	index := make([]byte, constants.HashLen)
	index[0] = 1

	fee := uint256.Zero()
	rawData := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, iat, numEpochs, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	utxo2 := &TXOut{}
	ds := &DataStore{}
	err = utxo2.NewDataStore(ds)
	assert.Nil(t, err)

	vin := []*TXIn{txin1}
	refUTXOs := []*TXOut{utxo1}
	vout := Vout{utxo2}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	currentHeight := constants.EpochLength*(iat+numEpochs) + 1
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if cleanup {
		t.Fatal("Should not be CleanupTx")
	}
}

func TestTxIsCleanupTxBad5(t *testing.T) {
	// Must have valid Vin and invalid Vout (bad ValueStore)
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	index := make([]byte, constants.HashLen)
	index[0] = 1

	fee := uint256.Zero()
	rawData := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, iat, numEpochs, fee)
	txin1, err := utxo1.MakeTxIn()
	assert.Nil(t, err)

	utxo2 := &TXOut{}
	vs := &ValueStore{}
	err = utxo2.NewValueStore(vs)
	assert.Nil(t, err)

	vin := []*TXIn{txin1}
	refUTXOs := []*TXOut{utxo1}
	vout := Vout{utxo2}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	currentHeight := constants.EpochLength*(iat+numEpochs) + 1
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if cleanup {
		t.Fatal("Should not be CleanupTx")
	}
}

func TestTxIsCleanupTxBad6(t *testing.T) {
	// Must have valid Vin and valid Vout; nonzero Fee
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	index := make([]byte, constants.HashLen)
	index[0] = 1

	fee := uint256.Zero()
	rawData := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, iat, numEpochs, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Compute remainingValue to have correct ValueStore
	currentHeight := constants.EpochLength*(iat+numEpochs) + 1
	remainingValue, err := utxo1.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 1, remainingValue, fee)

	vin := []*TXIn{txin1}
	refUTXOs := []*TXOut{utxo1}
	vout := Vout{utxo2}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.One(),
	}
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if cleanup {
		t.Fatal("Should not be CleanupTx")
	}
}

func TestTxIsCleanupTxBad7(t *testing.T) {
	// Must have valid Vin and valid Vout; nonzero Fee
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	index := make([]byte, constants.HashLen)
	index[0] = 1

	fee := uint256.Zero()
	rawData := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, iat, numEpochs, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Compute remainingValue to have correct ValueStore
	currentHeight := constants.EpochLength*(iat+numEpochs) + 1
	remainingValue, err := utxo1.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	newValue, err := new(uint256.Uint256).Add(remainingValue, uint256.One())
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 1, newValue, fee)

	vin := []*TXIn{txin1}
	refUTXOs := []*TXOut{utxo1}
	vout := Vout{utxo2}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if cleanup {
		t.Fatal("Should not be CleanupTx")
	}
}

func TestTxIsCleanupTxGood1(t *testing.T) {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	fee := uint256.Zero()
	index := make([]byte, constants.HashLen)
	index[0] = 1
	rawData := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData, index, iat, numEpochs, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Compute remainingValue to have correct ValueStore
	currentHeight := constants.EpochLength*(iat+numEpochs) + 1
	remainingValue, err := utxo1.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	utxo2 := makeVSWithValueFee(t, ownerSigner, 1, remainingValue, fee)

	vin := []*TXIn{txin1}
	refUTXOs := []*TXOut{utxo1}
	vout := []*TXOut{utxo2}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if !cleanup {
		t.Fatal("Should be valid CleanupTx")
	}
}

func TestTxIsCleanupTxGood2(t *testing.T) {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	fee := uint256.Zero()
	index1 := make([]byte, constants.HashLen)
	index1[0] = 1
	rawData1 := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(1)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData1, index1, iat, numEpochs, fee)
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	index2 := make([]byte, constants.HashLen)
	index2[0] = 2
	rawData2 := make([]byte, 100)
	utxo2 := makeDSWithValueFee(t, ownerSigner, 0, rawData2, index2, iat, numEpochs, fee)
	txin2, err := utxo2.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}

	// Compute remainingValue to have correct ValueStore
	currentHeight := constants.EpochLength*(iat+numEpochs) + 1
	vin := []*TXIn{txin1, txin2}
	refUTXOs := Vout{utxo1, utxo2}
	remainingValue, err := refUTXOs.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	utxo3 := makeVSWithValueFee(t, ownerSigner, 1, remainingValue, fee)

	vout := []*TXOut{utxo3}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	cleanup := tx.IsCleanupTx(currentHeight, refUTXOs)
	if !cleanup {
		t.Fatal("Should be valid CleanupTx")
	}
}

func TestTxIsCleanupTxGood3(t *testing.T) {
	// This does a full test of validation logic;
	// these fees should not affect the validity of the cleanup transaction
	// because no fees apply in this case.
	msg := MakeMockStorageGetter()
	dsFeeBig := big.NewInt(100)
	msg.SetDataStoreEpochFee(dsFeeBig)
	vsFeeBig := big.NewInt(1000)
	msg.SetValueStoreFee(vsFeeBig)
	tfFeeBig := big.NewInt(10000)
	msg.SetMinTxFee(tfFeeBig)
	msg.SetDataStoreEpochFee(dsFeeBig)
	storage := MakeStorage(msg)

	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	dsFee, err := new(uint256.Uint256).FromBigInt(dsFeeBig)
	if err != nil {
		t.Fatal(err)
	}
	index1 := make([]byte, constants.HashLen)
	index1[0] = 1
	rawData1 := make([]byte, 1)
	iat := uint32(1)
	numEpochs := uint32(3)
	utxo1 := makeDSWithValueFee(t, ownerSigner, 0, rawData1, index1, iat, numEpochs, dsFee)
	idx := uint32(0)
	err = utxo1.SetTxOutIdx(idx)
	if err != nil {
		t.Fatal(err)
	}
	txin1, err := utxo1.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo1.dataStore.Sign(txin1, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	index2 := make([]byte, constants.HashLen)
	index2[0] = 2
	rawData2 := make([]byte, 2)
	utxo2 := makeDSWithValueFee(t, ownerSigner, 0, rawData2, index2, iat, numEpochs, dsFee)
	idx = uint32(1)
	err = utxo2.SetTxOutIdx(idx)
	if err != nil {
		t.Fatal(err)
	}
	txin2, err := utxo2.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo2.dataStore.Sign(txin2, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	index3 := make([]byte, constants.HashLen)
	index3[0] = 3
	rawData3 := make([]byte, 3)
	utxo3 := makeDSWithValueFee(t, ownerSigner, 0, rawData3, index3, iat, numEpochs, dsFee)
	idx = uint32(2)
	err = utxo3.SetTxOutIdx(idx)
	if err != nil {
		t.Fatal(err)
	}
	txin3, err := utxo3.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	err = utxo3.dataStore.Sign(txin3, ownerSigner)
	if err != nil {
		t.Fatal(err)
	}

	// Compute remainingValue to have correct ValueStore
	cleanupFee := uint256.Zero()
	currentHeight := constants.EpochLength*(iat+numEpochs) + 1
	vin := []*TXIn{txin1, txin2, txin3}
	refUTXOs := Vout{utxo1, utxo2, utxo3}
	remainingValue, err := refUTXOs.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	utxo4 := makeVSWithValueFee(t, ownerSigner, 1, remainingValue, cleanupFee)

	vout := []*TXOut{utxo4}
	tx := &Tx{
		Vin:  vin,
		Vout: vout,
		Fee:  uint256.Zero(),
	}
	err = tx.SetTxHash()
	if err != nil {
		t.Fatal(err)
	}

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
}

func TestTxValidateEqualVinVoutBad1(t *testing.T) {
	currentHeight := uint32(0)
	refUTXOs := Vout{}
	tx := &Tx{}
	err := tx.ValidateEqualVinVout(currentHeight, refUTXOs)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateEqualVinVoutBad2(t *testing.T) {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	txin := &TXIn{}
	utxo := &TXOut{}

	currentHeight := uint32(1)
	refUTXOs := Vout{}
	tx := &Tx{
		Vin:  Vin{txin},
		Vout: Vout{utxo},
		Fee:  uint256.Zero(),
	}
	err := tx.ValidateEqualVinVout(currentHeight, refUTXOs)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestTxValidateEqualVinVoutBad3(t *testing.T) {
	ownerSigner := &crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}

	txin := &TXIn{}
	utxo := &TXOut{}
	vs := &ValueStore{}
	zero := uint256.Zero()
	one := uint256.One()
	vs.VSPreImage = &VSPreImage{}
	vs.VSPreImage.Value = one.Clone()
	vs.VSPreImage.Fee = zero.Clone()
	err := utxo.NewValueStore(vs)
	if err != nil {
		t.Fatal(err)
	}

	currentHeight := uint32(1)
	tx := &Tx{
		Vin:  Vin{txin},
		Vout: Vout{utxo},
		Fee:  uint256.Zero(),
	}
	err = tx.ValidateEqualVinVout(currentHeight, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}
