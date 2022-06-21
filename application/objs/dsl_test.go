package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

func TestDSLinkerGood(t *testing.T) {
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
	txHash := make([]byte, constants.HashLen)
	dsl := &DSLinker{
		DSPreImage: dsp,
		TxHash:     txHash,
	}
	dsl2 := &DSLinker{}
	dslBytes, err := dsl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = dsl2.UnmarshalBinary(dslBytes)
	if err != nil {
		t.Fatal(err)
	}
	dslEqual(t, dsl, dsl2)
}

func dslEqual(t *testing.T, dsl1, dsl2 *DSLinker) {
	dspi1 := dsl1.DSPreImage
	dspi2 := dsl2.DSPreImage
	dspiEqual(t, dspi1, dspi2)
	if !bytes.Equal(dsl1.TxHash, dsl2.TxHash) {
		t.Fatal("Do not agree on TxHash!")
	}
}

func TestDSLinkerBad1(t *testing.T) {
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
	txHash := make([]byte, constants.HashLen)
	dsl := &DSLinker{
		DSPreImage: dsp,
		TxHash:     txHash,
	}
	_, err = dsl.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise error for invalid DSPreImage!")
	}
}

func TestDSLinkerBad2(t *testing.T) {
	cid := uint32(2)
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
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, 31) // Invalid TxHash
	dsl := &DSLinker{
		DSPreImage: dsp,
		TxHash:     txHash,
	}
	dsl2 := &DSLinker{}
	dslBytes, err := dsl.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = dsl2.UnmarshalBinary(dslBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid TxHash: incorrect byte length!")
	}
}

func TestDSLinkerMarshalBinary(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	dsl := &DSLinker{}
	_, err = dsl.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestDSLinkerUnmarshalBinary(t *testing.T) {
	dsl := &DSLinker{}
	data := make([]byte, 0)
	err := dsl.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestDSLinkerPreHash(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.PreHash()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.PreHash()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestDSLinkerIssuedAt(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.IssuedAt()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.IssuedAt()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	dsl.DSPreImage = &DSPreImage{}
	iatTrue := uint32(1)
	dsl.DSPreImage.IssuedAt = iatTrue
	iat, err := dsl.IssuedAt()
	if err != nil {
		t.Fatal(err)
	}
	if iat != iatTrue {
		t.Fatal("IssuedAt is incorrect")
	}
}

func TestDSLinkerChainID(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.ChainID()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.ChainID()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	dsl.DSPreImage = &DSPreImage{}
	cidTrue := uint32(1)
	dsl.DSPreImage.ChainID = cidTrue
	cid, err := dsl.ChainID()
	if err != nil {
		t.Fatal(err)
	}
	if cid != cidTrue {
		t.Fatal("ChainID is incorrect")
	}
}

func TestDSLinkerIndex(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.Index()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.Index()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	dsl.DSPreImage = &DSPreImage{}
	_, err = dsl.Index()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	indexTrue := make([]byte, constants.HashLen)
	dsl.DSPreImage.Index = utils.CopySlice(indexTrue)
	index, err := dsl.Index()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(index, indexTrue) {
		t.Fatal("index is incorrect!")
	}
}

func TestDSLinkerOwner(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.Owner()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.Owner()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	dsl.DSPreImage = &DSPreImage{}
	dsl.DSPreImage.Owner = &DataStoreOwner{}
	_, err = dsl.Owner()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	owner := &Owner{}
	err = owner.New(acct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	err = dsl.DSPreImage.Owner.NewFromOwner(owner)
	if err != nil {
		t.Fatal(err)
	}
	_, err = dsl.Owner()
	if err != nil {
		t.Fatal(err)
	}
}

func TestDSLinkerRawData(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.RawData()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.RawData()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	dsl.DSPreImage = &DSPreImage{}
	rawDataTrue := make([]byte, constants.HashLen)
	dsl.DSPreImage.RawData = utils.CopySlice(rawDataTrue)
	rd, err := dsl.RawData()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(rawDataTrue, rd) {
		t.Fatal("Should be equal")
	}
}

func TestDSLinkerUTXOID(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.UTXOID()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.UTXOID()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	dsl.TxHash = make([]byte, constants.HashLen)
	dsl.DSPreImage = &DSPreImage{}
	dsl.DSPreImage.TXOutIdx = 0

	preUtxoID := make([]byte, constants.HashLen+4)
	utxoIDTrue := crypto.Hasher(preUtxoID)

	utxoID, err := dsl.UTXOID()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(utxoID, utxoIDTrue) {
		t.Fatal("utxo is incorrect!")
	}
}

func TestDSLinkerTxOutIdx(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.TxOutIdx()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.TxOutIdx()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	dsl.DSPreImage = &DSPreImage{}
	idx, err := dsl.TxOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	if idx != 0 {
		t.Fatal("Should be 0! (not set)")
	}
}

func TestDSLinkerSetTxOutIdx(t *testing.T) {
	idx := uint32(0)
	ds := &DataStore{}
	err := ds.DSLinker.SetTxOutIdx(idx)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	err = dsl.SetTxOutIdx(idx)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	dsl.DSPreImage = &DSPreImage{}
	err = dsl.SetTxOutIdx(idx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestDSLinkerIsExpired(t *testing.T) {
	ds := &DataStore{}
	currentHeight := uint32(1)
	_, err := ds.DSLinker.IsExpired(currentHeight)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.IsExpired(currentHeight)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestDSLinkerEpochOfExpiration(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.EpochOfExpiration()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.EpochOfExpiration()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestDSLinkerValue(t *testing.T) {
	ds := &DataStore{}
	_, err := ds.DSLinker.Value()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	_, err = dsl.Value()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestDSLinkerValidateSignature(t *testing.T) {
	ds := &DataStore{}
	currentHeight := uint32(0)
	msg := make([]byte, 0)
	sig := &DataStoreSignature{}
	err := ds.DSLinker.ValidateSignature(currentHeight, msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	err = dsl.ValidateSignature(currentHeight, msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestDSLinkerValidatePreSignature(t *testing.T) {
	ds := &DataStore{}
	msg := make([]byte, 0)
	sig := &DataStoreSignature{}
	err := ds.DSLinker.ValidatePreSignature(msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	dsl := &DSLinker{}
	err = dsl.ValidatePreSignature(msg, sig)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}
