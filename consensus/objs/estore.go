package objs

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"github.com/MadBase/MadNet/errorz"

	mdefs "github.com/MadBase/MadNet/consensus/objs/capn"
	"github.com/MadBase/MadNet/consensus/objs/estore"
	"github.com/MadBase/MadNet/interfaces"
	gUtils "github.com/MadBase/MadNet/utils"
	capnp "zombiezen.com/go/capnproto2"
)

// EncryptedStore ...
type EncryptedStore struct {
	cypherText []byte
	nonce      []byte
	Kid        []byte
	Name       []byte
	// Not Part of actual object below this line
	ClearText []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// EncryptedStore object
func (b *EncryptedStore) UnmarshalBinary(data []byte) error {
	bh, err := estore.Unmarshal(data)
	if err != nil {
		return err
	}
	return b.UnmarshalCapn(bh)
}

// UnmarshalCapn unmarshals the capnproto definition of the object
func (b *EncryptedStore) UnmarshalCapn(bh mdefs.EncryptedStore) error {
	err := estore.Validate(bh)
	if err != nil {
		return err
	}
	b.Name = gUtils.CopySlice(bh.Name())
	b.cypherText = gUtils.CopySlice(bh.CypherText())
	b.nonce = gUtils.CopySlice(bh.Nonce())
	b.Kid = gUtils.CopySlice(bh.Kid())
	return nil
}

// MarshalBinary takes the EncryptedStore object and returns the canonical
// byte slice
func (b *EncryptedStore) MarshalBinary() ([]byte, error) {
	bh, err := b.MarshalCapn(nil)
	if err != nil {
		return nil, err
	}
	err = estore.Validate(bh)
	if err != nil {
		return nil, err
	}
	return estore.Marshal(bh)
}

// MarshalCapn marshals the object into its capnproto definition
func (b *EncryptedStore) MarshalCapn(seg *capnp.Segment) (mdefs.EncryptedStore, error) {
	if b == nil {
		return mdefs.EncryptedStore{}, errorz.ErrInvalid{}.New("not initialized")
	}
	var bh mdefs.EncryptedStore
	if seg == nil {
		_, seg, err := capnp.NewMessage(capnp.SingleSegment(nil))
		if err != nil {
			return mdefs.EncryptedStore{}, err
		}
		tmp, err := mdefs.NewRootEncryptedStore(seg)
		if err != nil {
			return mdefs.EncryptedStore{}, err
		}
		bh = tmp
	} else {
		tmp, err := mdefs.NewRootEncryptedStore(seg)
		if err != nil {
			return mdefs.EncryptedStore{}, err
		}
		bh = tmp
	}
	name := gUtils.CopySlice(b.Name)
	err := bh.SetName(name)
	if err != nil {
		return mdefs.EncryptedStore{}, err
	}
	err = bh.SetCypherText(b.cypherText)
	if err != nil {
		return mdefs.EncryptedStore{}, err
	}
	err = bh.SetNonce(b.nonce)
	if err != nil {
		return mdefs.EncryptedStore{}, err
	}
	err = bh.SetKid(b.Kid)
	if err != nil {
		return mdefs.EncryptedStore{}, err
	}
	return bh, nil
}

// Encrypt encrypts b.ClearText and writes the result to b.cypherText;
// afterwards, it zeros b.ClearText and sets its pointer to nil
func (b *EncryptedStore) Encrypt(resolver interfaces.KeyResolver) error {
	key, err := resolver.GetKey(b.Kid)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	b.nonce = nonce
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	b.cypherText = aesgcm.Seal(nil, b.nonce, b.ClearText, nil)
	for k := 0; k < len(b.ClearText); k++ {
		b.ClearText[k] = 0
	}
	b.ClearText = nil
	return nil
}

// Decrypt decrypts b.cypherText and saves the result to b.ClearText
func (b *EncryptedStore) Decrypt(resolver interfaces.KeyResolver) error {
	key, err := resolver.GetKey(b.Kid)
	if err != nil {
		return err
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	plaintext, err := aesgcm.Open(nil, b.nonce, b.cypherText, nil)
	if err != nil {
		return err
	}
	b.ClearText = plaintext
	return nil
}
