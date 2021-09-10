package localrpc

import (
	"errors"

	from "github.com/MadBase/MadNet/application/objs"
	to "github.com/MadBase/MadNet/proto"
)

// TODO: ForwardTranslateTXOut and ReverseTranslateTXOut need to be updated
//		 to include TxFee option in switch case. There is no option at this
//		 point.
//
//		 The challenge is that there is nothing in particular that should
//		 require this information to be fully transmitted, as anyone is
//		 able to make a related TxFee object. This is because Tx objects
//		 will eventually be transmitted and will *include* TxFee objects.
//		 How this is specified as compactly as possible should be investigated.

func ForwardTranslateDataStore(f *from.DataStore) (*to.DataStore, error) {
	t := &to.DataStore{}
	if f == nil {
		return nil, errors.New("dataStore object should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	if f.DSLinker != nil {
		newDSLinker, err := ForwardTranslateDSLinker(f.DSLinker)
		if err != nil {
			return nil, err
		}
		t.DSLinker = newDSLinker
	}
	if f.Signature != nil {
		ownerBytes, err := f.Signature.MarshalBinary()
		if err != nil {
			return nil, err
		}
		newOwner, err := ForwardTranslateByte(ownerBytes)
		if err != nil {
			return nil, err
		}
		t.Signature = newOwner
	}
	return t, nil
}

func ForwardTranslateValueStore(f *from.ValueStore) (*to.ValueStore, error) {
	t := &to.ValueStore{}
	if f == nil {
		return nil, errors.New("valueStore object should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	newTxHash, err := ForwardTranslateByte(f.TxHash)
	if err != nil {
		return nil, err
	}
	t.TxHash = newTxHash
	if f.VSPreImage != nil {
		newVSPreImage, err := ForwardTranslateVSPreImage(f.VSPreImage)
		if err != nil {
			return nil, err
		}
		t.VSPreImage = newVSPreImage
	}
	return t, nil
}

func ForwardTranslateVSPreImage(f *from.VSPreImage) (*to.VSPreImage, error) {
	t := &to.VSPreImage{}
	if f == nil {
		return nil, errors.New("object of type VSPreImage should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	t.ChainID = f.ChainID
	if f.Owner != nil {
		ownerBytes, err := f.Owner.MarshalBinary()
		if err != nil {
			return nil, err
		}
		newOwner, err := ForwardTranslateByte(ownerBytes)
		if err != nil {
			return nil, err
		}
		t.Owner = newOwner
	}
	t.TXOutIdx = f.TXOutIdx
	t.Value, err = f.Value.MarshalString()
	if err != nil {
		return nil, err
	}
	t.Fee, err = f.Fee.MarshalString()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func ForwardTranslateASPreImage(f *from.ASPreImage) (*to.ASPreImage, error) {
	t := &to.ASPreImage{}
	if f == nil {
		return nil, errors.New("object of type ASPreImage should not be nil")
	}
	_, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	t.ChainID = f.ChainID
	t.Exp = f.Exp
	t.IssuedAt = f.IssuedAt
	if f.Owner != nil {
		ownerBytes, err := f.Owner.MarshalBinary()
		if err != nil {
			return nil, err
		}
		newOwner, err := ForwardTranslateByte(ownerBytes)
		if err != nil {
			return nil, err
		}
		t.Owner = newOwner
	}
	t.TXOutIdx = f.TXOutIdx
	t.Value, err = f.Value.MarshalString()
	if err != nil {
		return nil, err
	}
	t.Fee, err = f.Fee.MarshalString()
	if err != nil {
		return nil, err
	}
	return t, nil
}

func ForwardTranslateTXInLinker(f *from.TXInLinker) (*to.TXInLinker, error) {
	t := &to.TXInLinker{}
	if f == nil {
		return nil, errors.New("object of type TXInLinker should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	if f.TXInPreImage != nil {
		newTXInPreImage, err := ForwardTranslateTXInPreImage(f.TXInPreImage)
		if err != nil {
			return nil, err
		}
		t.TXInPreImage = newTXInPreImage
	}
	newTxHash, err := ForwardTranslateByte(f.TxHash)
	if err != nil {
		return nil, err
	}
	t.TxHash = newTxHash
	return t, nil
}

func ForwardTranslateTXOut(f *from.TXOut) (*to.TXOut, error) {
	if f == nil {
		return nil, errors.New("object of type TXOut should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	switch {
	case f.HasAtomicSwap():
		obj, err := f.AtomicSwap()
		if err != nil {
			return nil, err
		}
		newObj, err := ForwardTranslateAtomicSwap(obj)
		if err != nil {
			return nil, err
		}
		tt := &to.TXOut_AtomicSwap{AtomicSwap: newObj}
		t := &to.TXOut{Utxo: tt}
		return t, nil
	case f.HasValueStore():
		obj, err := f.ValueStore()
		if err != nil {
			return nil, err
		}
		newObj, err := ForwardTranslateValueStore(obj)
		if err != nil {
			return nil, err
		}
		tt := &to.TXOut_ValueStore{ValueStore: newObj}
		t := &to.TXOut{Utxo: tt}
		return t, nil
	case f.HasDataStore():
		obj, err := f.DataStore()
		if err != nil {
			return nil, err
		}
		newObj, err := ForwardTranslateDataStore(obj)
		if err != nil {
			return nil, err
		}
		tt := &to.TXOut_DataStore{DataStore: newObj}
		t := &to.TXOut{Utxo: tt}
		return t, nil
	case f.HasTxFee():
		obj, err := f.TxFee()
		if err != nil {
			return nil, err
		}
		newObj, err := ForwardTranslateTxFee(obj)
		if err != nil {
			return nil, err
		}
		tt := &to.TXOut_TxFee{TxFee: newObj}
		t := &to.TXOut{Utxo: tt}
		return t, nil
	default:
		return nil, errors.New("no txout in forward translate")
	}
}

func ForwardTranslateTx(f *from.Tx) (*to.Tx, error) {
	t := &to.Tx{}
	if f == nil {
		return nil, errors.New("tx object should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	for _, txIn := range f.Vin {
		newVin, err := ForwardTranslateTXIn(txIn)
		if err != nil {
			return nil, err
		}
		t.Vin = append(t.Vin, newVin)
	}
	for _, txOut := range f.Vout {
		newVout, err := ForwardTranslateTXOut(txOut)
		if err != nil {
			return nil, err
		}
		t.Vout = append(t.Vout, newVout)
	}
	return t, nil
}

func ForwardTranslateAtomicSwap(f *from.AtomicSwap) (*to.AtomicSwap, error) {
	t := &to.AtomicSwap{}
	if f == nil {
		return nil, errors.New("atomicSwap object should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	if f.ASPreImage != nil {
		newASPreImage, err := ForwardTranslateASPreImage(f.ASPreImage)
		if err != nil {
			return nil, err
		}
		t.ASPreImage = newASPreImage
	}
	newTxHash, err := ForwardTranslateByte(f.TxHash)
	if err != nil {
		return nil, err
	}
	t.TxHash = newTxHash
	return t, nil
}

func ForwardTranslateDSPreImage(f *from.DSPreImage) (*to.DSPreImage, error) {
	t := &to.DSPreImage{}
	if f == nil {
		return nil, errors.New("object of type DSPreImage should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	t.ChainID = f.ChainID
	t.Deposit, err = f.Deposit.MarshalString()
	if err != nil {
		return nil, err
	}
	t.Fee, err = f.Fee.MarshalString()
	if err != nil {
		return nil, err
	}
	newIndex, err := ForwardTranslateByte(f.Index)
	if err != nil {
		return nil, err
	}
	t.Index = newIndex
	t.IssuedAt = f.IssuedAt
	if f.Owner != nil {
		ownerBytes, err := f.Owner.MarshalBinary()
		if err != nil {
			return nil, err
		}
		newOwner, err := ForwardTranslateByte(ownerBytes)
		if err != nil {
			return nil, err
		}
		t.Owner = newOwner
	}
	newRawData, err := ForwardTranslateByte(f.RawData)
	if err != nil {
		return nil, err
	}
	t.RawData = newRawData
	t.TXOutIdx = f.TXOutIdx
	return t, nil
}

func ForwardTranslateTXInPreImage(f *from.TXInPreImage) (*to.TXInPreImage, error) {
	t := &to.TXInPreImage{}
	if f == nil {
		return nil, errors.New("object of type TXInPreImage should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	t.ChainID = f.ChainID
	newConsumedTxHash, err := ForwardTranslateByte(f.ConsumedTxHash)
	if err != nil {
		return nil, err
	}
	t.ConsumedTxHash = newConsumedTxHash
	t.ConsumedTxIdx = f.ConsumedTxIdx
	return t, nil
}

func ForwardTranslateDSLinker(f *from.DSLinker) (*to.DSLinker, error) {
	t := &to.DSLinker{}
	if f == nil {
		return nil, errors.New("object of type DSLinker should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	if f.DSPreImage != nil {
		newDSPreImage, err := ForwardTranslateDSPreImage(f.DSPreImage)
		if err != nil {
			return nil, err
		}
		t.DSPreImage = newDSPreImage
	}
	newTxHash, err := ForwardTranslateByte(f.TxHash)
	if err != nil {
		return nil, err
	}
	t.TxHash = newTxHash
	return t, nil
}

func ForwardTranslateTXIn(f *from.TXIn) (*to.TXIn, error) {
	t := &to.TXIn{}
	if f == nil {
		return nil, errors.New("object of type TXIn should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	newSignature, err := ForwardTranslateByte(f.Signature)
	if err != nil {
		return nil, err
	}
	t.Signature = newSignature
	if f.TXInLinker != nil {
		newTXInLinker, err := ForwardTranslateTXInLinker(f.TXInLinker)
		if err != nil {
			return nil, err
		}
		t.TXInLinker = newTXInLinker
	}
	return t, nil
}

func ForwardTranslateTxFee(f *from.TxFee) (*to.TxFee, error) {
	t := &to.TxFee{}
	if f == nil {
		return nil, errors.New("txFee object should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	newTxHash, err := ForwardTranslateByte(f.TxHash)
	if err != nil {
		return nil, err
	}
	t.TxHash = newTxHash
	if f.TFPreImage != nil {
		newTFPreImage, err := ForwardTranslateTFPreImage(f.TFPreImage)
		if err != nil {
			return nil, err
		}
		t.TFPreImage = newTFPreImage
	}
	return t, nil
}

func ForwardTranslateTFPreImage(f *from.TFPreImage) (*to.TFPreImage, error) {
	t := &to.TFPreImage{}
	if f == nil {
		return nil, errors.New("object of type TFPreImage should not be nil")
	}
	b, err := f.MarshalBinary()
	if err != nil {
		return nil, err
	}
	err = f.UnmarshalBinary(b)
	if err != nil {
		return nil, err
	}
	t.ChainID = f.ChainID
	t.TXOutIdx = f.TXOutIdx
	t.Fee, err = f.Fee.MarshalString()
	if err != nil {
		return nil, err
	}
	return t, nil
}
