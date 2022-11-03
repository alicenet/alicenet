package crypto

import (
	"crypto/ecdsa"
	"fmt"

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
		pubk, err = tryEthPersonalSign(msg, sig)
		if err != nil {
			return nil, err
		}
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

func tryEthPersonalSign(msg []byte, sig []byte) (*ecdsa.PublicKey, error) {
	ethMsg := []byte("\x19Ethereum Signed Message:\n" + fmt.Sprint(len(msg)))
	ethMsg = append(ethMsg, msg...)
	hsh := Hasher(ethMsg)
	// https://github.com/ethereum/go-ethereum/blob/master/signer/core/signed_data.go#L263
	if sig[64] != 27 && sig[64] != 28 {
		return nil, fmt.Errorf("V is not {0,1,2,3} || {27, 28})")
	}
	sig[64] -= 27 // Transform yellow paper V from 27/28 to 0/1
	pubk, err := eth.SigToPub(hsh, sig)
	if err != nil {
		return nil, err
	}
	return pubk, nil
}
