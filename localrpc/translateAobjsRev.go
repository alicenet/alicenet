package localrpc

import (
	"errors"

	to "github.com/alicenet/alicenet/application/objs"
	"github.com/alicenet/alicenet/application/objs/uint256"
	from "github.com/alicenet/alicenet/proto"
)

func ReverseTranslateDataStore(f *from.DataStore) (*to.DataStore, error) {
	t := &to.DataStore{}
	if f.DSLinker != nil {
		newDSLinker, err := ReverseTranslateDSLinker(f.DSLinker)
		if err != nil {
			return nil, err
		}
		t.DSLinker = newDSLinker
	}

	if f.Signature != "" {
		signatureBytes, err := ReverseTranslateByte(f.Signature)
		if err != nil {
			return nil, err
		}
		Signature := &to.DataStoreSignature{}
		err = Signature.UnmarshalBinary(signatureBytes)
		if err != nil {
			return nil, err
		}
		t.Signature = Signature
	}
	return t, nil
}

func ReverseTranslateValueStore(f *from.ValueStore) (*to.ValueStore, error) {
	t := &to.ValueStore{}
	newTxHash, err := ReverseTranslateByte(f.TxHash)
	if err != nil {
		return nil, err
	}

	t.TxHash = newTxHash

	if f.VSPreImage != nil {
		newVSPreImage, err := ReverseTranslateVSPreImage(f.VSPreImage)
		if err != nil {
			return nil, err
		}
		t.VSPreImage = newVSPreImage
	}

	return t, nil
}

func ReverseTranslateVSPreImage(f *from.VSPreImage) (*to.VSPreImage, error) {
	t := &to.VSPreImage{}
	t.ChainID = f.ChainID

	if f.Owner != "" {
		ownerBytes, err := ReverseTranslateByte(f.Owner)
		if err != nil {
			return nil, err
		}
		newOwner := &to.ValueStoreOwner{}
		err = newOwner.UnmarshalBinary(ownerBytes)
		if err != nil {
			return nil, err
		}
		t.Owner = newOwner
	}

	t.TXOutIdx = f.TXOutIdx

	t.Value = &uint256.Uint256{}
	err := t.Value.UnmarshalString(f.Value)
	if err != nil {
		return nil, err
	}
	if len(f.Fee) == 0 {
		f.Fee = "0"
	}
	t.Fee = &uint256.Uint256{}
	err = t.Fee.UnmarshalString(f.Fee)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func ReverseTranslateASPreImage(f *from.ASPreImage) (*to.ASPreImage, error) {
	t := &to.ASPreImage{}
	t.ChainID = f.ChainID
	t.Exp = f.Exp
	t.IssuedAt = f.IssuedAt

	if f.Owner != "" {
		ownerBytes, err := ReverseTranslateByte(f.Owner)
		if err != nil {
			return nil, err
		}
		newOwner := &to.AtomicSwapOwner{}
		err = newOwner.UnmarshalBinary(ownerBytes)
		if err != nil {
			return nil, err
		}
		t.Owner = newOwner
	}

	t.TXOutIdx = f.TXOutIdx

	t.Value = &uint256.Uint256{}
	err := t.Value.UnmarshalString(f.Value)
	if err != nil {
		return nil, err
	}
	if len(f.Fee) == 0 {
		f.Fee = "0"
	}
	t.Fee = &uint256.Uint256{}
	err = t.Fee.UnmarshalString(f.Fee)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func ReverseTranslateTXInLinker(f *from.TXInLinker) (*to.TXInLinker, error) {
	t := &to.TXInLinker{}
	if f.TXInPreImage != nil {
		newTXInPreImage, err := ReverseTranslateTXInPreImage(f.TXInPreImage)
		if err != nil {
			return nil, err
		}
		t.TXInPreImage = newTXInPreImage
	}

	newTxHash, err := ReverseTranslateByte(f.TxHash)
	if err != nil {
		return nil, err
	}

	t.TxHash = newTxHash
	return t, nil
}

func ReverseTranslateTx(f *from.Tx) (*to.Tx, error) {
	t := &to.Tx{}
	for _, txIn := range f.Vin {
		newVin, err := ReverseTranslateTXIn(txIn)
		if err != nil {
			return nil, err
		}
		t.Vin = append(t.Vin, newVin)
	}

	for _, txOut := range f.Vout {
		newVout, err := ReverseTranslateTXOut(txOut)
		if err != nil {
			return nil, err
		}
		t.Vout = append(t.Vout, newVout)
	}

	if len(f.Fee) == 0 {
		f.Fee = "0"
	}
	t.Fee = &uint256.Uint256{}
	err := t.Fee.UnmarshalString(f.Fee)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func ReverseTranslateTXOut(f *from.TXOut) (*to.TXOut, error) {
	t := &to.TXOut{}
	switch f.GetUtxo().(type) {
	case *from.TXOut_AtomicSwap:
		ff := f.GetAtomicSwap()
		obj, err := ReverseTranslateAtomicSwap(ff)
		if err != nil {
			return nil, err
		}

		err = t.NewAtomicSwap(obj)
		if err != nil {
			return nil, err
		}
	case *from.TXOut_ValueStore:
		ff := f.GetValueStore()
		obj, err := ReverseTranslateValueStore(ff)
		if err != nil {
			return nil, err
		}

		err = t.NewValueStore(obj)
		if err != nil {
			return nil, err
		}
	case *from.TXOut_DataStore:
		ff := f.GetDataStore()
		obj, err := ReverseTranslateDataStore(ff)
		if err != nil {
			return nil, err
		}

		err = t.NewDataStore(obj)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid")
	}
	_, err := t.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func ReverseTranslateAtomicSwap(f *from.AtomicSwap) (*to.AtomicSwap, error) {
	t := &to.AtomicSwap{}
	if f.ASPreImage != nil {
		newASPreImage, err := ReverseTranslateASPreImage(f.ASPreImage)
		if err != nil {
			return nil, err
		}
		t.ASPreImage = newASPreImage
	}

	newTxHash, err := ReverseTranslateByte(f.TxHash)
	if err != nil {
		return nil, err
	}
	t.TxHash = newTxHash
	return t, nil
}

func ReverseTranslateDSPreImage(f *from.DSPreImage) (*to.DSPreImage, error) {
	t := &to.DSPreImage{}
	t.ChainID = f.ChainID

	t.Deposit = &uint256.Uint256{}
	err := t.Deposit.UnmarshalString(f.Deposit)
	if err != nil {
		return nil, err
	}
	if len(f.Fee) == 0 {
		f.Fee = "0"
	}
	t.Fee = &uint256.Uint256{}
	err = t.Fee.UnmarshalString(f.Fee)
	if err != nil {
		return nil, err
	}

	newIndex, err := ReverseTranslateByte(f.Index)
	if err != nil {
		return nil, err
	}

	t.Index = newIndex
	t.IssuedAt = f.IssuedAt

	if f.Owner != "" {
		ownerBytes, err := ReverseTranslateByte(f.Owner)
		if err != nil {
			return nil, err
		}
		newOwner := &to.DataStoreOwner{}
		err = newOwner.UnmarshalBinary(ownerBytes)
		if err != nil {
			return nil, err
		}
		t.Owner = newOwner
	}

	newRawData, err := ReverseTranslateByte(f.RawData)
	if err != nil {
		return nil, err
	}

	t.RawData = newRawData
	t.TXOutIdx = f.TXOutIdx

	return t, nil
}

func ReverseTranslateTXInPreImage(f *from.TXInPreImage) (*to.TXInPreImage, error) {
	t := &to.TXInPreImage{}
	t.ChainID = f.ChainID

	newConsumedTxHash, err := ReverseTranslateByte(f.ConsumedTxHash)
	if err != nil {
		return nil, err
	}

	t.ConsumedTxHash = newConsumedTxHash
	t.ConsumedTxIdx = f.ConsumedTxIdx
	return t, nil
}

func ReverseTranslateDSLinker(f *from.DSLinker) (*to.DSLinker, error) {
	t := &to.DSLinker{}
	if f.DSPreImage != nil {
		newDSPreImage, err := ReverseTranslateDSPreImage(f.DSPreImage)
		if err != nil {
			return nil, err
		}
		t.DSPreImage = newDSPreImage
	}

	newTxHash, err := ReverseTranslateByte(f.TxHash)
	if err != nil {
		return nil, err
	}

	t.TxHash = newTxHash
	return t, nil
}

func ReverseTranslateTXIn(f *from.TXIn) (*to.TXIn, error) {
	t := &to.TXIn{}
	newSignature, err := ReverseTranslateByte(f.Signature)
	if err != nil {
		return nil, err
	}
	t.Signature = newSignature

	if f.TXInLinker != nil {
		newTXInLinker, err := ReverseTranslateTXInLinker(f.TXInLinker)
		if err != nil {
			return nil, err
		}
		t.TXInLinker = newTXInLinker
	}
	return t, nil
}
