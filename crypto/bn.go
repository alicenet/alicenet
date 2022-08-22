package crypto

import (
	"math/big"

	bn256 "github.com/alicenet/alicenet/crypto/bn256/cloudflare"
)

// BNSigner creates cryptographic signatures using the bn256 curve.
type BNSigner struct {
	privk *big.Int
	pubk  *bn256.G2
}

// SetPrivk sets the private key of the BNSigner.
func (bns *BNSigner) SetPrivk(privk []byte) error {
	if bns == nil {
		return ErrInvalid
	}
	bns.privk = new(big.Int).SetBytes(privk)
	bns.privk.Mod(bns.privk, bn256.Order)
	bns.pubk = new(bn256.G2).ScalarBaseMult(bns.privk)
	return nil
}

// Pubkey returns the marshalled public key of the BNSigner.
func (bns *BNSigner) Pubkey() ([]byte, error) {
	if bns == nil || bns.privk == nil {
		return nil, ErrPrivkNotSet
	}
	return bns.pubk.Marshal(), nil
}

// Sign will generate a signature for msg using the private key of the
// BNSigner.
func (bns *BNSigner) Sign(msg []byte) ([]byte, error) {
	if bns == nil || bns.privk == nil {
		return nil, ErrPrivkNotSet
	}
	pubkbytes := bns.pubk.Marshal()
	nmsg := []byte{}
	nmsg = append(nmsg, pubkbytes...)
	nmsg = append(nmsg, msg...)
	sigpoint, err := bn256.Sign(nmsg, bns.privk, bn256.HashToG1)
	if err != nil {
		return nil, err
	}
	return bn256.MarshalSignature(sigpoint, bns.pubk)
}

// BNValidator is the object that performs cryptographic validation of
// BNSigner signatures.
type BNValidator struct {
}

// Validate will validate a BNSigner signature for msg.
func (bnv *BNValidator) Validate(msg []byte, sig []byte) ([]byte, error) {
	pubkey, signature, err := bn256.UnmarshalSignature(sig)
	if err != nil {
		return nil, err
	}
	pubkbytes := pubkey.Marshal()
	nmsg := []byte{}
	nmsg = append(nmsg, pubkbytes...)
	nmsg = append(nmsg, msg...)
	ok, err := bn256.Verify(nmsg, signature, pubkey, bn256.HashToG1)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidSignature
	}
	return pubkbytes, nil
}

// PubkeyFromSig returns the public key of the signer from the signature.
func (bnv *BNValidator) PubkeyFromSig(sig []byte) ([]byte, error) {
	pubkey, _, err := bn256.UnmarshalSignature(sig)
	if err != nil {
		return nil, err
	}
	return pubkey.Marshal(), nil
}
