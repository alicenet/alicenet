package objs

/* // todo remove after AS holding logic implemented
import (
	"bytes"
	"testing"

	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
)

func TestAtomicSwapGood(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	priOwner := crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := crypto.Secp256k1Signer{}
	if err := altOwner.SetPrivk(crypto.Hasher([]byte("b"))); err != nil {
		t.Fatal(err)
	}
	altPubk, err := altOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	altOwnerAcct := crypto.GetAccount(altPubk)

	hashKey := crypto.Hasher([]byte("foo"))
	owner := &AtomicSwapOwner{}
	err = owner.New(priOwnerAcct, altOwnerAcct, hashKey)
	if err != nil {
		t.Fatal(err)
	}

	iat := uint32(1)
	exp := uint32(1234)
	asp := &ASPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		IssuedAt: iat,
		Exp:      exp,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen)
	as := &AtomicSwap{
		ASPreImage: asp,
		TxHash:     txHash,
	}
	as2 := &AtomicSwap{}
	asBytes, err := as.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = as2.UnmarshalBinary(asBytes)
	if err != nil {
		t.Fatal(err)
	}
	asEqual(t, as, as2)
}

func asEqual(t *testing.T, as1, as2 *AtomicSwap) {
	aspi1 := as1.ASPreImage
	aspi2 := as2.ASPreImage
	aspiEqual(t, aspi1, aspi2)
	if !bytes.Equal(as1.TxHash, as2.TxHash) {
		t.Fatal("Do not agree on TxHash!")
	}
}

func TestAtomicSwapBad1(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	cid := uint32(0) // Invalid ChainID
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	priOwner := crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := crypto.Secp256k1Signer{}
	if err := altOwner.SetPrivk(crypto.Hasher([]byte("b"))); err != nil {
		t.Fatal(err)
	}
	altPubk, err := altOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	altOwnerAcct := crypto.GetAccount(altPubk)

	hashKey := crypto.Hasher([]byte("foo"))
	owner := &AtomicSwapOwner{}
	err = owner.New(priOwnerAcct, altOwnerAcct, hashKey)
	if err != nil {
		t.Fatal(err)
	}

	iat := uint32(1)
	exp := uint32(1234)
	asp := &ASPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		IssuedAt: iat,
		Exp:      exp,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen)
	as := &AtomicSwap{
		ASPreImage: asp,
		TxHash:     txHash,
	}
	as2 := &AtomicSwap{}
	asBytes, err := as.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = as2.UnmarshalBinary(asBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid ASPreImage!")
	}
}

func TestAtomicSwapBad2(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	cid := uint32(2)
	val, err := new(uint256.Uint256).FromUint64(65537)
	if err != nil {
		t.Fatal(err)
	}
	txoid := uint32(17)

	priOwner := &crypto.Secp256k1Signer{}
	if err := priOwner.SetPrivk(crypto.Hasher([]byte("a"))); err != nil {
		t.Fatal(err)
	}
	priPubk, err := priOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	priOwnerAcct := crypto.GetAccount(priPubk)

	altOwner := &crypto.Secp256k1Signer{}
	if err := altOwner.SetPrivk(crypto.Hasher([]byte("b"))); err != nil {
		t.Fatal(err)
	}
	altPubk, err := altOwner.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	altOwnerAcct := crypto.GetAccount(altPubk)

	hashKey := crypto.Hasher([]byte("foo"))
	owner := &AtomicSwapOwner{}
	err = owner.New(priOwnerAcct, altOwnerAcct, hashKey)
	if err != nil {
		t.Fatal(err)
	}

	iat := uint32(1)
	exp := uint32(3)
	asp := &ASPreImage{
		ChainID:  cid,
		Value:    val,
		TXOutIdx: txoid,
		Owner:    owner,
		IssuedAt: iat,
		Exp:      exp,
		Fee:      new(uint256.Uint256).SetZero(),
	}
	txHash := make([]byte, constants.HashLen+1) // Invalid TxHash
	as := &AtomicSwap{
		ASPreImage: asp,
		TxHash:     txHash,
	}
	as2 := &AtomicSwap{}
	asBytes, err := as.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	err = as2.UnmarshalBinary(asBytes)
	if err == nil {
		t.Fatal("Should raise error for invalid TxHash: incorrect byte length!")
	}

	asValue, err := as.Value()
	if err != nil {
		t.Fatal(err)
	}
	if val != asValue {
		t.Fatal("as.Next does not agree")
	}
	asExp, err := as.Exp()
	if err != nil {
		t.Fatal(err)
	}
	if exp != asExp {
		t.Fatal("as.Exp does not agree")
	}
	txOutIdx, err := as.TxOutIdx()
	if err != nil {
		t.Fatal(err)
	}
	if txoid != txOutIdx {
		t.Fatal("as.TXOutIdex does not agree")
	}
	asChainID, err := as.ChainID()
	if err != nil {
		t.Fatal(err)
	}
	if cid != asChainID {
		t.Fatal("as.ChainID does not agree")
	}
	asIssuedAt, err := as.IssuedAt()
	if err != nil {
		t.Fatal(err)
	}
	if iat != asIssuedAt {
		t.Fatal("as.IssuedAt does not agree")
	}
	asOwner, err := as.Owner()
	if err != nil {
		t.Fatal(err)
	}
	if owner != asOwner {
		t.Fatal("as.Owner does not agree")
	}

	utxoid, err := as.UTXOID()
	if err != nil {
		t.Fatal(err)
	}
	utxoid2, err := as.UTXOID()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(utxoid, utxoid2) {
		t.Fatal("as.UTXOIDs fail to agree")
	}

	ph, err := as.PreHash()
	if err != nil {
		t.Fatal(err)
	}
	ph2, err := as.ASPreImage.PreHash()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ph, ph2) {
		t.Fatal("PreHashes fail to agree")
	}
	isExpired, err := as.IsExpired(constants.EpochLength)
	if err != nil {
		t.Fatal(err)
	}
	if isExpired {
		t.Fatal("Should not be expired")
	}

	asTxHash := make([]byte, constants.HashLen)
	err = as.SetTxHash(asTxHash)
	if err != nil {
		t.Fatal(err)
	}
	txIn, err := as.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
	txIn.TXInLinker.TxHash = crypto.Hasher([]byte{})
	err = as.SignAsPrimary(txIn, priOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = as.ValidateSignature(constants.EpochLength*(exp+1), txIn)
	if err != nil {
		t.Fatal(err)
	}

	err = as.SignAsPrimary(txIn, priOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = as.ValidateSignature(constants.EpochLength, txIn)
	if err == nil {
		t.Fatal("ValSig should fail")
	}

	err = as.SignAsAlternate(txIn, altOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = as.ValidateSignature(constants.EpochLength, txIn)
	if err != nil {
		t.Fatal(err)
	}

	err = as.SignAsAlternate(txIn, altOwner, crypto.Hasher(hashKey))
	if err == nil {
		t.Fatal("SignAsAlt should fail")
	}
	as.ASPreImage.Owner.HashLock = crypto.Hasher(crypto.Hasher(hashKey))
	err = as.ValidateSignature(constants.EpochLength, txIn)
	if err == nil {
		t.Fatal(err)
	}

	as.ASPreImage.Owner.HashLock = crypto.Hasher(hashKey)
	err = as.SignAsAlternate(txIn, altOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = as.ValidateSignature(constants.EpochLength*(exp+1), txIn)
	if err == nil {
		t.Fatal("Should raise error in ValSig")
	}

	err = as.SignAsAlternate(txIn, priOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = as.ValidateSignature(constants.EpochLength*(exp+1), txIn)
	if err == nil {
		t.Fatal("Should raise error in ValSig")
	}
	err = as.SignAsAlternate(txIn, priOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = as.ValidateSignature(constants.EpochLength, txIn)
	if err == nil {
		t.Fatal("Should raise error in ValSig")
	}

	err = as.SignAsPrimary(txIn, altOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = as.ValidateSignature(constants.EpochLength*(exp+1), txIn)
	if err == nil {
		t.Fatal("Should raise error in ValSig")
	}
	err = as.SignAsPrimary(txIn, altOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	err = as.ValidateSignature(constants.EpochLength, txIn)
	if err == nil {
		t.Fatal("Should raise error in ValSig")
	}

	isExpired, err = as.IsExpired(constants.EpochLength * (exp + 1))
	if err != nil {
		t.Fatal(err)
	}
	if !isExpired {
		t.Fatal("Should be expired")
	}
}

func TestAtomicSwapMarshalBinary(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	as := &AtomicSwap{}
	_, err = as.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	as.ASPreImage = &ASPreImage{}
	_, err = as.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
}

func TestAtomicSwapPreHash(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.PreHash()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	_, err = as.PreHash()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestAtomicSwapUTXOID(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.UTXOID()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	_, err = as.UTXOID()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestAtomicSwapTxOutIdx(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.TxOutIdx()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	_, err = as.TxOutIdx()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestAtomicSwapSetTxOutIdx(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	idx := uint32(0)
	utxo := &TXOut{}
	err := utxo.atomicSwap.SetTxOutIdx(idx)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	err = as.SetTxOutIdx(idx)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	as.ASPreImage = &ASPreImage{}
	err = as.SetTxOutIdx(idx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAtomicSwapSetTxHash(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	txHash := make([]byte, 0)
	utxo := &TXOut{}
	err := utxo.atomicSwap.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	err = as.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	as.ASPreImage = &ASPreImage{}
	err = as.SetTxHash(txHash)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	txHash = make([]byte, constants.HashLen)
	err = as.SetTxHash(txHash)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAtomicSwapValue(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.Value()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	_, err = as.Value()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestAtomicSwapOwner(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.Owner()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	as := &AtomicSwap{}
	_, err = as.Owner()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	as.ASPreImage = &ASPreImage{}
	as.ASPreImage.Owner = &AtomicSwapOwner{}
	_, err = as.Owner()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	curveSpec := constants.CurveSecp256k1
	priAcct := make([]byte, constants.OwnerLen)
	priOwner := &Owner{}
	err = priOwner.New(priAcct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	altAcct := make([]byte, constants.OwnerLen)
	altOwner := &Owner{}
	err = altOwner.New(altAcct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	hashKey := make([]byte, constants.HashLen)
	err = as.ASPreImage.Owner.NewFromOwner(priOwner, altOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	_, err = as.Owner()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAtomicSwapGenericOwner(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.GenericOwner()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	as := &AtomicSwap{}
	_, err = as.GenericOwner()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	as.ASPreImage = &ASPreImage{}
	as.ASPreImage.Owner = &AtomicSwapOwner{}
	curveSpec := constants.CurveSecp256k1
	priAcct := make([]byte, constants.OwnerLen)
	priOwner := &Owner{}
	err = priOwner.New(priAcct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	altAcct := make([]byte, constants.OwnerLen)
	altOwner := &Owner{}
	err = altOwner.New(altAcct, curveSpec)
	if err != nil {
		t.Fatal(err)
	}
	hashKey := make([]byte, constants.HashLen)
	err = as.ASPreImage.Owner.NewFromOwner(priOwner, altOwner, hashKey)
	if err != nil {
		t.Fatal(err)
	}
	_, err = as.GenericOwner()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestAtomicSwapChainID(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.ChainID()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	_, err = as.ChainID()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	as.ASPreImage = &ASPreImage{}
	cid := uint32(1)
	as.ASPreImage.ChainID = cid
	_, err = as.ChainID()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAtomicSwapExp(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.Exp()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	_, err = as.Exp()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestAtomicSwapIssuedAt(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.IssuedAt()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	_, err = as.IssuedAt()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestAtomicSwapIsExpired(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	currentHeight := constants.EpochLength
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.IsExpired(currentHeight)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	_, err = as.IsExpired(currentHeight)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestAtomicSwapValidateSignature(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	txIn := &TXIn{}
	currentHeight := uint32(0)
	utxo := &TXOut{}
	err := utxo.atomicSwap.ValidateSignature(currentHeight, txIn)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	as := &AtomicSwap{}
	err = as.ValidateSignature(currentHeight, txIn)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

}

func TestAtomicSwapSigning(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	txIn := &TXIn{}
	signer := &crypto.Secp256k1Signer{}
	hashKey := make([]byte, constants.HashLen)
	utxo := &TXOut{}
	err := utxo.atomicSwap.SignAsPrimary(txIn, signer, hashKey)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	err = as.SignAsPrimary(txIn, signer, hashKey)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	err = utxo.atomicSwap.SignAsAlternate(txIn, signer, hashKey)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	err = as.SignAsAlternate(txIn, signer, hashKey)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	currentHeight := uint32(0)
	err = as.ValidateSignature(currentHeight, txIn)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}
}

func TestAtomicSwapMakeTxIn(t *testing.T) {
	// todo remove after AS holding logic implemented
	return
	utxo := &TXOut{}
	_, err := utxo.atomicSwap.MakeTxIn()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	as := &AtomicSwap{}
	_, err = as.MakeTxIn()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	as.ASPreImage = &ASPreImage{}
	txOutIdx := uint32(1)
	as.ASPreImage.TXOutIdx = txOutIdx
	_, err = as.MakeTxIn()
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	cid := uint32(1)
	as.ASPreImage.ChainID = cid
	_, err = as.MakeTxIn()
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
	as.TxHash = make([]byte, constants.HashLen)
	_, err = as.MakeTxIn()
	if err != nil {
		t.Fatal(err)
	}
}

func TestAtomicSwapValidateFee(t *testing.T) {
	return
	panic("test not implemented")
}
*/
