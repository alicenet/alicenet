package objs

import (
	"bytes"
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
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

	vsp := &VSPreImage{
		ChainID: cid,
		Value:   val,
		Owner:   owner,
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

func TestTx(t *testing.T) {

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
	err = tx.ValidateEqualVinVout(consumedUTXOs, 1)
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
	_, err = tx2.Validate(nil, 1, consumedUTXOs)
	if err != nil {
		t.Fatal(err)
	}

	txVec := TxVec([]*Tx{tx})
	err = txVec.Validate(1, consumedUTXOs)
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
	signer.SetPrivk(privk)
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

func TestTxConsumedPreHash(t *testing.T) {
	tx := &Tx{}
	_, err := tx.ConsumedPreHash()
	if err == nil {
		t.Fatal("Should raise an error")
	}
}
