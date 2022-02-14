package estore

import (
	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/errorz"
	capnp "github.com/MadBase/go-capnproto2/v2"
)

// Marshal will marshal the EncryptedStore object.
func Marshal(v mdefs.EncryptedStore) ([]byte, error) {
	raw, err := capnp.Canonicalize(v.Struct)
	if err != nil {
		return nil, err
	}
	defer v.Struct.Segment().Message().Reset(nil)
	return raw, nil
}

// Unmarshal will unmarshal the EncryptedStore object.
func Unmarshal(data []byte) (mdefs.EncryptedStore, error) {
	var err error
	fn := func() (mdefs.EncryptedStore, error) {
		defer func() {
			if r := recover(); r != nil {
				err = errorz.ErrInvalid{}.New("bad serialization")
			}
		}()
		msg := &capnp.Message{Arena: capnp.SingleSegment(data)}
		obj, tmp := mdefs.ReadRootEncryptedStore(msg)
		err = tmp
		return obj, err
	}
	obj, err := fn()
	if err != nil {
		return mdefs.EncryptedStore{}, err
	}
	return obj, nil
}

// Validate will validate the EncryptedStore object
func Validate(p mdefs.EncryptedStore) error {
	if !p.IsValid() {
		return errorz.ErrInvalid{}.New("enc store capn obj is not valid")
	}
	if !p.HasCypherText() {
		return errorz.ErrInvalid{}.New("enc store capn obj does not have CypherText")
	}
	if !p.HasNonce() {
		return errorz.ErrInvalid{}.New("enc store capn obj does not have Nonce")
	}
	if !p.HasKid() {
		return errorz.ErrInvalid{}.New("enc store capn obj does not have Kid")
	}
	if !p.HasName() {
		return errorz.ErrInvalid{}.New("enc store capn obj does not have Name")
	}
	return nil
}
