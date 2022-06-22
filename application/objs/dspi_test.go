package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
)

func TestDSPreImageGood(t *testing.T) {
	cid := uint32(2)
	idxPre := []byte("Index")
	idx := crypto.Hasher(idxPre)
	txoid := uint32(17)
	iat := uint32(1)
	rawdata := make([]byte, 1)
	dep, err := new(uint256.Uint256).FromUint64(uint64((len(rawdata) + constants.BaseDatasizeConst) * 3))
	if err != nil {
		t.Fatal(err)
	}

	ownerSigner := crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &DataStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	dsp := &DSPreImage{
		ChainID:  cid,
		Index:    idx,
		IssuedAt: iat,
		Deposit:  dep,
		RawData:  rawdata,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	dsp2 := &DSPreImage{}
	dspBytes, err := dsp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = dsp2.UnmarshalBinary(dspBytes)
	if err != nil {
		t.Fatal(err)
	}
	dspiEqual(t, dsp, dsp2)
}

func dspiEqual(t *testing.T, dspi1, dspi2 *DSPreImage) {
	if dspi1.ChainID != dspi2.ChainID {
		t.Fatal("Do not agree on ChainID!")
	}
	if !bytes.Equal(dspi1.Index, dspi2.Index) {
		t.Fatal("Do not agree on Index!")
	}
	if dspi1.IssuedAt != dspi2.IssuedAt {
		t.Fatal("Do not agree on IssuedAt!")
	}
	if !dspi1.Deposit.Eq(dspi2.Deposit) {
		t.Fatal("Do not agree on Deposit!")
	}
	if !bytes.Equal(dspi1.RawData, dspi2.RawData) {
		t.Fatal("Do not agree on RawData!")
	}
	if dspi1.TXOutIdx != dspi2.TXOutIdx {
		t.Fatal("Do not agree on TXOutIdx!")
	}
	if dspi1.Owner.SVA != dspi2.Owner.SVA {
		t.Fatal("Do not agree on CurveSpec!")
	}
	if dspi1.Owner.CurveSpec != dspi2.Owner.CurveSpec {
		t.Fatal("Do not agree on CurveSpec!")
	}
	if !bytes.Equal(dspi1.Owner.Account, dspi2.Owner.Account) {
		t.Fatal("Do not agree on Index!")
	}
}

func TestDSPreImageBad1(t *testing.T) {
	cid := uint32(0) // Invalid ChainID
	idxPre := []byte("Index")
	idx := crypto.Hasher(idxPre)
	txoid := uint32(17)
	iat := uint32(1)
	rawdata := crypto.Hasher([]byte("foo"))
	dep, err := new(uint256.Uint256).FromUint64(uint64((len(rawdata) + constants.BaseDatasizeConst) * 3))
	if err != nil {
		t.Fatal(err)
	}

	ownerSigner := crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &DataStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	dsp := &DSPreImage{
		ChainID:  cid,
		Index:    idx,
		IssuedAt: iat,
		Deposit:  dep,
		RawData:  rawdata,
		TXOutIdx: txoid,
		Owner:    owner,
	}
	_, err = dsp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise error for invalid ChainID!")
	}
}

func TestDSPreImageBad2(t *testing.T) {
	cid := uint32(2)
	idx := []byte("Index") // Invalid Index
	txoid := uint32(17)
	iat := uint32(1)
	rawdata := crypto.Hasher([]byte("foo"))
	dep, err := new(uint256.Uint256).FromUint64(uint64((len(rawdata) + constants.BaseDatasizeConst) * 3))
	if err != nil {
		t.Fatal(err)
	}

	ownerSigner := crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &DataStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	dsp := &DSPreImage{
		ChainID:  cid,
		Index:    idx,
		IssuedAt: iat,
		Deposit:  dep,
		RawData:  rawdata,
		TXOutIdx: txoid,
		Owner:    owner,
	}
	_, err = dsp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise error for invalid Index!")
	}
}

func TestDSPreImageBad3(t *testing.T) {
	cid := uint32(2)
	idxPre := []byte("Index")
	idx := crypto.Hasher(idxPre)
	txoid := uint32(17)
	iat := uint32(0) // Invalid IssuedAt
	rawdata := crypto.Hasher([]byte("foo"))
	dep, err := new(uint256.Uint256).FromUint64(uint64((len(rawdata) + constants.BaseDatasizeConst) * 3))
	if err != nil {
		t.Fatal(err)
	}

	ownerSigner := crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &DataStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	dsp := &DSPreImage{
		ChainID:  cid,
		Index:    idx,
		IssuedAt: iat,
		Deposit:  dep,
		RawData:  rawdata,
		TXOutIdx: txoid,
		Owner:    owner,
	}
	_, err = dsp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise error for invalid IssuedAt!")
	}
}

func TestDSPreImageBad4(t *testing.T) {
	cid := uint32(2)
	idxPre := []byte("Index")
	idx := crypto.Hasher(idxPre)
	txoid := uint32(17)
	iat := uint32(1)
	rawdata := crypto.Hasher([]byte("foo"))
	dep, err := new(uint256.Uint256).FromUint64(uint64(len(rawdata) + constants.BaseDatasizeConst)) // Invalid Deposit
	if err != nil {
		t.Fatal(err)
	}

	ownerSigner := crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &DataStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	dsp := &DSPreImage{
		ChainID:  cid,
		Index:    idx,
		IssuedAt: iat,
		Deposit:  dep,
		RawData:  rawdata,
		TXOutIdx: txoid,
		Owner:    owner,
	}
	_, err = dsp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise error for invalid Deposit!")
	}
}

func TestDSPreImageBad5(t *testing.T) {
	cid := uint32(2)
	idxPre := []byte("Index")
	idx := crypto.Hasher(idxPre)
	txoid := uint32(17)
	iat := uint32(1)
	rawdata := make([]byte, 32)
	dep, err := new(uint256.Uint256).FromUint64(uint64((len(rawdata) + constants.BaseDatasizeConst) * 3))
	if err != nil {
		t.Fatal(err)
	}

	ownerSigner := crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &DataStoreOwner{}
	owner.New(ownerAcct, 0)

	dsp := &DSPreImage{
		ChainID:  cid,
		Index:    idx,
		IssuedAt: iat,
		Deposit:  dep,
		RawData:  rawdata,
		TXOutIdx: txoid,
		Owner:    owner,
	}
	_, err = dsp.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise error for invalid Owner!")
	}
}

func TestDSOwnerSig(t *testing.T) {
	cid := uint32(2)
	idxPre := []byte("Index")
	idx := crypto.Hasher(idxPre)
	txoid := uint32(17)
	iat := uint32(1)
	rawdata := make([]byte, 32)
	dep, err := new(uint256.Uint256).FromUint64(uint64((len(rawdata) + constants.BaseDatasizeConst) * 3))
	if err != nil {
		t.Fatal(err)
	}

	ownerSigner := crypto.Secp256k1Signer{}
	if err := ownerSigner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	ownerPubk, err := ownerSigner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	ownerAcct := crypto.GetAccount(ownerPubk)
	owner := &DataStoreOwner{}
	owner.New(ownerAcct, constants.CurveSecp256k1)

	dsp := &DSPreImage{
		ChainID:  cid,
		Index:    idx,
		IssuedAt: iat,
		Deposit:  dep,
		RawData:  rawdata,
		TXOutIdx: txoid,
		Owner:    owner,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	_, err = dsp.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDSPreImageMarshalBinary(t *testing.T) {
	dsl := &DSLinker{}
	_, err := dsl.DSPreImage.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsp := &DSPreImage{}
	_, err = dsp.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestDSPreImageUnmarshalBinary(t *testing.T) {
	dsp := &DSPreImage{}
	data := make([]byte, 0)
	err := dsp.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestDSPreImagePreHash(t *testing.T) {
	dsl := &DSLinker{}
	_, err := dsl.DSPreImage.PreHash()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsp := &DSPreImage{}
	_, err = dsp.PreHash()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	rawData := make([]byte, constants.HashLen)
	dataSize := uint32(len(rawData))
	numEpochs := uint32(5)
	iat := uint32(1)
	cid := uint32(1)
	index := crypto.Hasher([]byte("Test"))
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	dso := &DataStoreOwner{}
	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	dso.New(acct, curveSpec)
	txOutIdx := uint32(1)

	dsp.ChainID = cid
	dsp.Index = index
	dsp.IssuedAt = iat
	dsp.Deposit = deposit
	dsp.RawData = rawData
	dsp.TXOutIdx = txOutIdx
	dsp.Owner = dso
	dsp.Fee = new(uint256.Uint256).SetZero()

	_, err = dsp.PreHash()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDSPreImageRemainingValue(t *testing.T) {
	dsl := &DSLinker{}
	currentHeight := uint32(0)
	_, err := dsl.DSPreImage.RemainingValue(currentHeight)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsp := &DSPreImage{}

	rawData := make([]byte, constants.HashLen)
	dataSize := uint32(len(rawData))
	numEpochs := uint32(5)
	iat := uint32(2) // This needs to be > 1 for test
	cid := uint32(1)
	index := crypto.Hasher([]byte("Test"))
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	dso := &DataStoreOwner{}
	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	dso.New(acct, curveSpec)
	txOutIdx := uint32(1)

	dsp.ChainID = cid
	dsp.Index = index
	dsp.IssuedAt = iat
	dsp.Deposit = deposit
	dsp.RawData = rawData
	dsp.TXOutIdx = txOutIdx
	dsp.Owner = dso

	// Next during epoch of creation
	currentHeight = iat * constants.EpochLength
	remVal, err := dsp.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	if !remVal.Eq(deposit) {
		t.Fatal("Incorrect value (1)")
	}

	// Next in epoch *before* creation
	currentHeight = (iat - 1) * constants.EpochLength
	remVal, err = dsp.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	if !remVal.Eq(deposit) {
		t.Fatal("Incorrect value (2)")
	}

	epochCost, err := new(uint256.Uint256).FromUint64(uint64(dataSize + constants.BaseDatasizeConst))
	if err != nil {
		t.Fatal(err)
	}

	eoe, err := dsp.EpochOfExpiration()
	if err != nil {
		t.Fatal(err)
	}

	// Next in epoch right before expiration
	currentHeight = (eoe - 1) * constants.EpochLength
	remVal, err = dsp.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	remValTrue, err := new(uint256.Uint256).Add(epochCost, epochCost)
	if err != nil {
		t.Fatal(err)
	}
	if !remVal.Eq(remValTrue) {
		t.Fatal("Incorrect value (3)")
	}

	// Next in epoch of expiration
	currentHeight = eoe * constants.EpochLength
	remVal, err = dsp.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	remValTrue = epochCost.Clone()
	if !remVal.Eq(remValTrue) {
		t.Fatal("Incorrect value (4)")
	}

	// Next in epoch following epoch of expiration
	currentHeight = (eoe + 1) * constants.EpochLength
	remVal, err = dsp.RemainingValue(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	remValTrue = epochCost.Clone()
	if !remVal.Eq(remValTrue) {
		t.Fatal("Incorrect value (5)")
	}
}

func TestDSPreImageValue(t *testing.T) {
	dsl := &DSLinker{}
	_, err := dsl.DSPreImage.Value()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsp := &DSPreImage{}
	_, err = dsp.Value()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	dataSize := uint32(1)
	numEpochs := uint32(1)
	depositTrue, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	dsp.Deposit = depositTrue
	deposit, err := dsp.Value()
	if err != nil {
		t.Fatal(err)
	}
	if !depositTrue.Eq(deposit) {
		t.Fatal("Should not happen!")
	}
}

func TestDSPreImageValidatePreSignature(t *testing.T) {
	dsl := &DSLinker{}
	msg := make([]byte, 0)
	sig := &DataStoreSignature{}
	sig.SVA = DataStoreSVA
	sig.CurveSpec = constants.CurveSecp256k1
	sig.Signature = make([]byte, constants.CurveSecp256k1SigLen)
	err := dsl.DSPreImage.ValidatePreSignature(msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsp := &DSPreImage{}
	err = dsp.ValidatePreSignature(msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	rawData := crypto.Hasher([]byte("RawData"))
	dataSize := uint32(len(rawData))
	numEpochs := uint32(10)
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	dsp.RawData = rawData
	dsp.Deposit = deposit
	dso := &DataStoreOwner{}
	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	dso.New(acct, curveSpec)
	dsp.Owner = dso
	err = dsp.ValidatePreSignature(msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestDSPreImageValidateSignature(t *testing.T) {
	dsl := &DSLinker{}
	currentHeight := uint32(0)
	msg := make([]byte, 0)
	sig := &DataStoreSignature{}
	sig.SVA = DataStoreSVA
	sig.CurveSpec = constants.CurveSecp256k1
	sig.Signature = make([]byte, constants.CurveSecp256k1SigLen)
	err := dsl.DSPreImage.ValidateSignature(currentHeight, msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsp := &DSPreImage{}
	err = dsp.ValidateSignature(currentHeight, msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	currentHeight = uint32(1)
	rawData := crypto.Hasher([]byte("RawData"))
	dataSize := uint32(len(rawData))
	numEpochs := uint32(10)
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	dsp.RawData = rawData
	dsp.Deposit = deposit
	dso := &DataStoreOwner{}
	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	dso.New(acct, curveSpec)
	dsp.Owner = dso
	err = dsp.ValidateSignature(currentHeight, msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestDSPreImageIsExpired(t *testing.T) {
	dsl := &DSLinker{}
	currentHeight := uint32(1)
	_, err := dsl.DSPreImage.IsExpired(currentHeight)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	dsp := &DSPreImage{}
	_, err = dsp.IsExpired(currentHeight)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	rawData := make([]byte, constants.HashLen)
	dataSize := uint32(len(rawData))
	numEpochs := uint32(5)
	iat := uint32(1)
	cid := uint32(1)
	index := crypto.Hasher([]byte("Test"))
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	dso := &DataStoreOwner{}
	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	dso.New(acct, curveSpec)
	txOutIdx := uint32(1)

	dsp.ChainID = cid
	dsp.Index = index
	dsp.IssuedAt = iat
	dsp.Deposit = deposit
	dsp.RawData = rawData
	dsp.TXOutIdx = txOutIdx
	dsp.Owner = dso

	eoe, err := dsp.EpochOfExpiration()
	if err != nil {
		t.Fatal(err)
	}

	currentHeight = eoe * constants.EpochLength
	expBool, err := dsp.IsExpired(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	if !expBool {
		t.Fatal("Should be expired!")
	}

	currentHeight = (eoe - 1) * constants.EpochLength
	expBool, err = dsp.IsExpired(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	if expBool {
		t.Fatal("Should not be expired!")
	}
}

func TestDSPreImageEpochOfExpiration(t *testing.T) {
	dsl := &DSLinker{}
	_, err := dsl.DSPreImage.EpochOfExpiration()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	dsp := &DSPreImage{}
	_, err = dsp.EpochOfExpiration()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	rawData := make([]byte, constants.HashLen)
	dataSize := uint32(len(rawData))
	numEpochs := uint32(5)
	iat := uint32(1)
	cid := uint32(1)
	index := crypto.Hasher([]byte("Test"))
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	dso := &DataStoreOwner{}
	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	dso.New(acct, curveSpec)
	txOutIdx := uint32(1)

	dsp.ChainID = cid
	dsp.Index = index
	dsp.IssuedAt = iat
	dsp.Deposit = uint256.One() // will lead to error
	dsp.RawData = rawData
	dsp.TXOutIdx = txOutIdx
	dsp.Owner = dso

	_, err = dsp.EpochOfExpiration()
	if err == nil {
		// Deposit is too small and raises error in NumEpochsEquation
		t.Fatal("Should have raised error (3)")
	}

	dsp.Deposit = deposit
	eoe, err := dsp.EpochOfExpiration()
	if err != nil {
		t.Fatal(err)
	}

	currentHeight := eoe * constants.EpochLength
	expBool, err := dsp.IsExpired(currentHeight)
	if err != nil {
		t.Fatal(err)
	}
	if !expBool {
		t.Fatal("Should be expired!")
	}
}

func TestDSPreImageValidateDeposit(t *testing.T) {
	dsl := &DSLinker{}
	err := dsl.DSPreImage.ValidateDeposit()
	if err == nil {
		// Fails because dsp == nil
		t.Fatal("Should raise an error (1)")
	}

	dsp := &DSPreImage{}
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because ChainID == 0
		t.Fatal("Should raise an error (2)")
	}

	cid := uint32(1)
	dsp.ChainID = cid
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because len(Index) != constants.HashLen
		t.Fatal("Should raise an error (3)")
	}

	index := crypto.Hasher([]byte("Index"))
	dsp.Index = index
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because IssuedAt == 0
		t.Fatal("Should raise an error (4)")
	}

	iat := uint32(1)
	dsp.IssuedAt = iat
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because len(rawData) == 0
		t.Fatal("Should raise an error (5)")
	}

	tooLargeData := constants.MaxDataStoreSize + 1
	rawData := make([]byte, tooLargeData)
	dsp.RawData = rawData
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because len(rawData) is too large
		t.Fatal("Should raise an error (6)")
	}

	rawData = make([]byte, 1)
	dsp.RawData = rawData
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because error in NumEpochsEquation (because Deposit == 0)
		t.Fatal("Should raise an error (7)")
	}

	numEpochs := uint32(0)
	dataSize := uint32(len(rawData))
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	dsp.Deposit = deposit
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because numEpochs == 0
		t.Fatal("Should raise an error (8)")
	}

	dsp.Deposit, err = new(uint256.Uint256).FromUint64(uint64(constants.MaxUint32))
	if err != nil {
		t.Fatal(err)
	}
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because computed deposit does not match dsp.Deposit
		t.Fatal("Should raise an error (9)")
	}

	numEpochs = uint32(1)
	deposit, err = BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	dsp.Deposit = deposit
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because dsp.Fee == nil
		t.Fatal("Should raise an error (7)")
	}

	dsp.Fee = new(uint256.Uint256)
	err = dsp.ValidateDeposit()
	if err == nil {
		// Fails because dsp.Owner == nil
		t.Fatal("Should raise an error (8)")
	}

	dso := &DataStoreOwner{}
	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	dso.New(acct, curveSpec)
	dsp.Owner = dso
	err = dsp.ValidateDeposit()
	if err != nil {
		// Should succeed
		t.Fatal(err)
	}
}

func TestBaseDepositEquation(t *testing.T) {
	dataSize := uint32(0)
	numEpochs := uint32(0)
	val, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	valTrue, err := new(uint256.Uint256).FromUint64(uint64(uint32(constants.BaseDatasizeConst) * uint32(2)))
	if err != nil {
		t.Fatal(err)
	}
	if !val.Eq(valTrue) {
		t.Fatal("Incorrect value")
	}

	// Too large dataSize
	dataSize = constants.MaxDataStoreSize + 1
	_, err = BaseDepositEquation(dataSize, numEpochs)
	if err == nil {
		t.Fatal("Should raise an error as dataSize < MaxDataStoreSize")
	}

	/*
		// Test no longer applies because no overflow should occur
		// uint32 conversion should fail
		dataSize = constants.MaxDataStoreSize
		numEpochs = uint32(1) << 31
		_, err = BaseDepositEquation(dataSize, numEpochs)
		if err == nil {
			t.Fatal("Should raise an error as uint32 conversion should fail")
		}
	*/

	dataSize = 1
	numEpochs = 1
	val, err = BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	valTrue, err = new(uint256.Uint256).FromUint64(uint64(uint32(constants.BaseDatasizeConst+1) * uint32(3)))
	if err != nil {
		t.Fatal(err)
	}
	if !val.Eq(valTrue) {
		t.Fatal("Incorrect value")
	}
}

func TestNumEpochsEquation(t *testing.T) {
	dataSize := uint32(1)
	numEpochsTrue := uint32(1)
	depositTrue, err := BaseDepositEquation(dataSize, numEpochsTrue)
	if err != nil {
		t.Fatal(err)
	}
	numEpochs, err := NumEpochsEquation(dataSize, depositTrue)
	if err != nil {
		t.Fatal(err)
	}
	if numEpochs != numEpochsTrue {
		t.Fatal("Incorrect numEpochs value")
	}

	dataSize = uint32(0)
	numEpochsTrue = uint32(0)
	depositTrue, err = BaseDepositEquation(dataSize, numEpochsTrue)
	if err != nil {
		t.Fatal(err)
	}
	numEpochs, err = NumEpochsEquation(dataSize, depositTrue)
	if err != nil {
		t.Fatal(err)
	}
	if numEpochs != numEpochsTrue {
		t.Fatal("Incorrect numEpochs value (2)")
	}

	dataSize = constants.MaxDataStoreSize + 1
	_, err = NumEpochsEquation(dataSize, depositTrue)
	if err == nil {
		t.Fatal("Should raise an error for dataSize too large")
	}

	dataSize = uint32(1)
	depositFalse, err := new(uint256.Uint256).FromUint64(uint64(constants.BaseDatasizeConst))
	if err != nil {
		t.Fatal(err)
	}
	_, err = NumEpochsEquation(dataSize, depositFalse)
	if err == nil {
		t.Fatal("Should raise an error for integer overflow")
	}
}

func TestRewardDepositEquationGood1(t *testing.T) {
	dataSize := uint32(1)
	numEpochs := uint32(3)
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	epochCost, err := new(uint256.Uint256).FromUint64(uint64(dataSize + constants.BaseDatasizeConst))
	if err != nil {
		t.Fatal(err)
	}

	epochInitial := uint32(1)
	epochFinal := uint32(1) // epochFinal == epochInitial, so should have full value
	tmp, err := new(uint256.Uint256).FromUint64(uint64(epochFinal - 1))
	if err != nil {
		t.Fatal(err)
	}
	tmp, err = new(uint256.Uint256).Mul(tmp, epochCost)
	if err != nil {
		t.Fatal(err)
	}
	remTrue, err := new(uint256.Uint256).Sub(deposit, tmp)
	if err != nil {
		t.Fatal(err)
	}
	remainder, err := RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}
	if !remainder.Eq(deposit) {
		t.Fatal("remainder should be full deposit")
	}

	epochFinal = uint32(2)
	tmp, err = new(uint256.Uint256).FromUint64(uint64(epochFinal - 1))
	if err != nil {
		t.Fatal(err)
	}
	tmp, err = new(uint256.Uint256).Mul(tmp, epochCost)
	if err != nil {
		t.Fatal(err)
	}
	remTrue, err = new(uint256.Uint256).Sub(deposit, tmp)
	if err != nil {
		t.Fatal(err)
	}
	remainder, err = RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}

	epochFinal = uint32(3)
	tmp, err = new(uint256.Uint256).FromUint64(uint64(epochFinal - 1))
	if err != nil {
		t.Fatal(err)
	}
	tmp, err = new(uint256.Uint256).Mul(tmp, epochCost)
	if err != nil {
		t.Fatal(err)
	}
	remTrue, err = new(uint256.Uint256).Sub(deposit, tmp)
	if err != nil {
		t.Fatal(err)
	}
	remainder, err = RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}

	epochFinal = uint32(4) // == epochInitial + numEpochs
	tmp, err = new(uint256.Uint256).FromUint64(uint64(epochFinal - 1))
	if err != nil {
		t.Fatal(err)
	}
	tmp, err = new(uint256.Uint256).Mul(tmp, epochCost)
	if err != nil {
		t.Fatal(err)
	}
	remTrue, err = new(uint256.Uint256).Sub(deposit, tmp)
	if err != nil {
		t.Fatal(err)
	}
	remainder, err = RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}

	epochFinal = uint32(5)                                             // == epochInitial + numEpochs + 1
	tmp, err = new(uint256.Uint256).FromUint64(uint64(epochFinal - 1)) // This is when DataStore should be cleaned up, so RemainingValue is epochCost
	if err != nil {
		t.Fatal(err)
	}
	tmp, err = new(uint256.Uint256).Mul(tmp, epochCost)
	if err != nil {
		t.Fatal(err)
	}
	remTrue, err = new(uint256.Uint256).Sub(deposit, tmp)
	if err != nil {
		t.Fatal(err)
	}
	remainder, err = RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}

	epochFinal = uint32(6) // == epochInitial + numEpochs + 2
	remTrue = epochCost    // We are past the required epochs, so RemainingValue is still epochCost
	// We should never get to this point because it should be cleaned up at
	// this point, but this would still be the reward if it was cleaned up later.
	remainder, err = RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}
}

func TestRewardDepositEquationGood2(t *testing.T) {
	dataSize := uint32(1)
	numEpochs := uint32(10)
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}
	epochCost, err := new(uint256.Uint256).FromUint64(uint64(dataSize + constants.BaseDatasizeConst))
	if err != nil {
		t.Fatal(err)
	}

	epochInitial := uint32(1)
	epochFinal := epochInitial // epochFinal == epochInitial, so should have full value
	tmp, err := new(uint256.Uint256).FromUint64(uint64(epochFinal - 1))
	if err != nil {
		t.Fatal(err)
	}
	tmp, err = new(uint256.Uint256).Mul(tmp, epochCost)
	if err != nil {
		t.Fatal(err)
	}
	remTrue, err := new(uint256.Uint256).Sub(deposit, tmp)
	if err != nil {
		t.Fatal(err)
	}
	remainder, err := RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}
	if !remainder.Eq(deposit) {
		t.Fatal("remainder should be full deposit")
	}

	// Last epoch where DataStore is valid
	epochFinal = epochInitial + numEpochs
	remTrue, err = new(uint256.Uint256).Add(epochCost, epochCost)
	if err != nil {
		t.Fatal(err)
	}
	remainder, err = RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}

	// DataStore is in EpochOfExpiration; it is now expired
	epochFinal = epochInitial + numEpochs + 1
	remTrue = epochCost.Clone() // This is when DataStore should be cleaned up, so RemainingValue is epochCost
	remainder, err = RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}

	// DataStore is past EpochOfExpiration; it is expired
	epochFinal = epochInitial + numEpochs + 2
	remTrue = epochCost // We are past the required epochs, so RemainingValue is still epochCost
	// We should never get to this point because it should be cleaned up at
	// this point, but this would still be the reward if it was cleaned up later.
	remainder, err = RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err != nil {
		t.Fatal(err)
	}
	if !remainder.Eq(remTrue) {
		t.Fatal("Invalid remainder for epoch =", epochFinal)
	}
}
func TestRewardDepositEquationBad(t *testing.T) {
	dataSize := uint32(1)
	numEpochs := uint32(3)
	deposit, err := BaseDepositEquation(dataSize, numEpochs)
	if err != nil {
		t.Fatal(err)
	}

	epochInitial := uint32(1)
	epochFinal := uint32(0) // Should not happen
	_, err = RewardDepositEquation(deposit, dataSize, epochInitial, epochFinal)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	depositBad := uint256.Zero()
	epochFinal = uint32(1)
	_, err = RewardDepositEquation(depositBad, dataSize, epochInitial, epochFinal)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
}
