package crypto

import (
	"crypto/ecdsa"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
	eth "github.com/ethereum/go-ethereum/crypto"
)

// Secp256k1Signer creates cryptographic signatures using the secp256k1 curve.
type Secp256k1Signer struct {
	privk *ecdsa.PrivateKey
	pubk  []byte
}

// Pubkey returns the marshalled public key of the Secp256k1Signer
// (uncompressed format).
func (secps *Secp256k1Signer) Pubkey() ([]byte, error) {
	if secps == nil || secps.privk == nil {
		return nil, ErrPrivkNotSet
	}
	return utils.CopySlice(secps.pubk), nil
}

// SetPrivk sets the private key of the Secp256k1Signer;
// privk is required to be 32 bytes!
func (secps *Secp256k1Signer) SetPrivk(privk []byte) error {
	if secps == nil {
		return ErrInvalid
	}
	ecprivk, err := eth.ToECDSA(privk)
	if err != nil {
		return err
	}
	secps.privk = ecprivk
	pubk := eth.FromECDSAPub(&ecprivk.PublicKey)
	secps.pubk = pubk
	return nil
}

// Sign will generate a signature for msg using the private key of the
// Secp256k1Signer; eth.Sign *assumes* we are signing the
// *hash of the message* (digestHash) and *not* the message itself.
func (secps *Secp256k1Signer) Sign(msg []byte) ([]byte, error) {
	if secps == nil || secps.privk == nil {
		return nil, ErrPrivkNotSet
	}
	digestHash := Hasher(msg)
	return eth.Sign(digestHash, secps.privk)
}

// Secp256k1Validator is a struct which allows for validation of cryptographic
// signatures from Secp256k1Signer.
type Secp256k1Validator struct {
}

// Validate will validate a Secp256k1Signer signature for msg.
func (secpv *Secp256k1Validator) Validate(msg []byte, sig []byte) ([]byte, error) {
	if len(sig) != constants.CurveSecp256k1SigLen {
		return nil, ErrInvalidSignature
	}
	hsh := Hasher(msg)
	pubk, err := eth.SigToPub(hsh, sig)
	if err != nil {
		return nil, err
	}
	pubkbytes := eth.FromECDSAPub(pubk)
	return pubkbytes, nil
}

// PubkeyFromSig returns the public key of the signer from the signature
// and message.
func (secpv *Secp256k1Validator) PubkeyFromSig(msg []byte, sig []byte) ([]byte, error) {
	if len(sig) != constants.CurveSecp256k1SigLen {
		return nil, ErrInvalidSignature
	}
	hsh := Hasher(msg)
	pubk, err := eth.SigToPub(hsh, sig)
	if err != nil {
		return nil, err
	}
	pubkbytes := eth.FromECDSAPub(pubk)
	return pubkbytes, nil
}
