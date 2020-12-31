package objs

import (
	mdefs "github.com/MadBase/MadNet/application/objs/capn"
	"github.com/MadBase/MadNet/application/objs/datastore"
	"github.com/MadBase/MadNet/application/objs/uint256"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// DataStore is a datastore UTXO
type DataStore struct {
	DSLinker  *DSLinker
	Signature *DataStoreSignature
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// DataStore object
func (b *DataStore) UnmarshalBinary(data []byte) error {
	bc, err := datastore.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bc)
}

// MarshalBinary takes the DataStore object and returns the canonical
// byte slice
func (b *DataStore) MarshalBinary() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	bc, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	return datastore.Marshal(bc)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *DataStore) UnmarshalCapn(bc mdefs.DataStore) error {
	if err := datastore.Validate(bc); err != nil {
		return err
	}
	dSLinker := &DSLinker{}
	if err := dSLinker.UnmarshalCapn(bc.DSLinker()); err != nil {
		return err
	}
	b.DSLinker = dSLinker
	sig := &DataStoreSignature{}
	if err := sig.UnmarshalBinary(bc.Signature()); err != nil {
		return err
	}
	b.Signature = sig
	return nil
}

// MarshalCapn marshals the object into its capnproto definition
func (b *DataStore) MarshalCapn(seg *capnp.Segment) (mdefs.DataStore, error) {
	if b == nil {
		return mdefs.DataStore{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bc mdefs.DataStore
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return bc, err
		}
		tmp, err := mdefs.NewRootDataStore(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	} else {
		tmp, err := mdefs.NewDataStore(seg)
		if err != nil {
			return bc, err
		}
		bc = tmp
	}
	seg = bc.Struct.Segment()
	bt, err := b.DSLinker.MarshalCapn(seg)
	if err != nil {
		return bc, err
	}
	if err := bc.SetDSLinker(bt); err != nil {
		return bc, err
	}
	sig, err := b.Signature.MarshalBinary()
	if err != nil {
		return bc, err
	}
	if err := bc.SetSignature(sig); err != nil {
		return bc, err
	}
	return bc, nil
}

// IssuedAt returns the IssuedAt of the object
func (b *DataStore) IssuedAt() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.IssuedAt()
}

// ChainID returns the ChainID of the object
func (b *DataStore) ChainID() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.ChainID()
}

// Index returns the Index of the object
func (b *DataStore) Index() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.Index()
}

// PreHash returns the PreHash of the object
func (b *DataStore) PreHash() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.PreHash()
}

// UTXOID returns the UTXOID of the object
func (b *DataStore) UTXOID() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.UTXOID()
}

// TXOutIdx returns the TXOutIdx of the object
func (b *DataStore) TXOutIdx() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.TXOutIdx()
}

// SetTXOutIdx sets the TXOutIdx of the object
func (b *DataStore) SetTXOutIdx(idx uint32) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.SetTXOutIdx(idx)
}

// TxHash returns the TxHash of the object
func (b *DataStore) TxHash() ([]byte, error) {
	if b == nil || b.DSLinker == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	if len(b.DSLinker.TxHash) != constants.HashLen {
		return nil, errorz.ErrInvalid{}.New("invalid TxHash")
	}
	return utils.CopySlice(b.DSLinker.TxHash), nil
}

// SetTxHash sets the TxHash of the object
func (b *DataStore) SetTxHash(txHash []byte) error {
	if b == nil || b.DSLinker == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	if len(txHash) != constants.HashLen {
		return errorz.ErrInvalid{}.New("Invalid hash length")
	}
	b.DSLinker.TxHash = utils.CopySlice(txHash)
	return nil
}

// RawData returns the RawData field of the sub object
func (b *DataStore) RawData() ([]byte, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.RawData()
}

// Owner returns the DataStoreOwner of the DataStore
func (b *DataStore) Owner() (*DataStoreOwner, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.Owner()
}

// GenericOwner returns the Owner of the DataStore
func (b *DataStore) GenericOwner() (*Owner, error) {
	dso, err := b.Owner()
	if err != nil {
		return nil, err
	}
	onr := &Owner{}
	err = onr.NewFromDataStoreOwner(dso)
	if err != nil {
		return nil, err
	}
	return onr, nil
}

// EpochOfExpiration returns the epoch in which the datastore may be garbage
// collected
func (b *DataStore) EpochOfExpiration() (uint32, error) {
	if b == nil {
		return 0, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.EpochOfExpiration()
}

// RemainingValue returns remaining value at the time of consumption
func (b *DataStore) RemainingValue(currentHeight uint32) (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.RemainingValue(currentHeight)
}

// Value returns the value stored in the object at the time of creation
func (b *DataStore) Value() (*uint256.Uint256, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	return b.DSLinker.Value()
}

// ValidatePreSignature validates the signature of the datastore at the time of
// creation
func (b *DataStore) ValidatePreSignature() error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	msg, err := b.DSLinker.MarshalBinary()
	if err != nil {
		return err
	}
	return b.DSLinker.ValidatePreSignature(msg, b.Signature)
}

// PreSign generates the signature for a DataStore at the time of creation
func (b *DataStore) PreSign(s Signer) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	msg, err := b.DSLinker.MarshalBinary()
	if err != nil {
		return err
	}
	owner, err := b.Owner()
	if err != nil {
		return err
	}
	sig, err := owner.Sign(msg, s)
	if err != nil {
		return err
	}
	b.Signature = sig
	return nil
}

// Sign generates the signature for a DataStore at the time of consumption
func (b *DataStore) Sign(txIn *TXIn, s Signer) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	msg, err := txIn.TXInLinker.MarshalBinary()
	if err != nil {
		return err
	}
	owner, err := b.Owner()
	if err != nil {
		return err
	}
	sig, err := owner.Sign(msg, s)
	if err != nil {
		return err
	}
	sigb, err := sig.MarshalBinary()
	if err != nil {
		return err
	}
	txIn.Signature = sigb
	return nil
}

// ValidateSignature validates the signature of the datastore at the time of
// consumption
func (b *DataStore) ValidateSignature(currentHeight uint32, txIn *TXIn) error {
	if b == nil {
		return errorz.ErrInvalid{}.New("not initialized")
	}
	msg, err := txIn.TXInLinker.MarshalBinary()
	if err != nil {
		return err
	}
	sig := &DataStoreSignature{}
	if err := sig.UnmarshalBinary(txIn.Signature); err != nil {
		return err
	}
	return b.DSLinker.ValidateSignature(currentHeight, msg, sig)
}

// MakeTxIn constructs a TXIn object for the current object
func (b *DataStore) MakeTxIn() (*TXIn, error) {
	if b == nil {
		return nil, errorz.ErrInvalid{}.New("not initialized")
	}
	txOutIdx, err := b.TXOutIdx()
	if err != nil {
		return nil, err
	}
	cid, err := b.ChainID()
	if err != nil {
		return nil, err
	}
	txHash, err := b.TxHash()
	if err != nil {
		return nil, err
	}
	return &TXIn{
		TXInLinker: &TXInLinker{
			TXInPreImage: &TXInPreImage{
				ConsumedTxIdx:  txOutIdx,
				ConsumedTxHash: utils.CopySlice(txHash),
				ChainID:        cid,
			},
		},
	}, nil
}
