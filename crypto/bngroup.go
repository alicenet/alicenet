package crypto

import (
	"math/big"

	"github.com/alicenet/alicenet/constants"
	bn256 "github.com/alicenet/alicenet/crypto/bn256/cloudflare"
)

// BNGroupSigner creates cryptographic signatures using the bn256 curve.
type BNGroupSigner struct {
	privk     *big.Int
	pubk      *bn256.G2
	groupPubk *bn256.G2
}

// SetPrivk sets the private key of the BNGroupSigner.
func (bns *BNGroupSigner) SetPrivk(privk []byte) error {
	if bns == nil {
		return ErrInvalid
	}
	bns.privk = new(big.Int).SetBytes(privk)
	bns.privk.Mod(bns.privk, bn256.Order)
	bns.pubk = new(bn256.G2).ScalarBaseMult(bns.privk)
	return nil
}

// SetGroupPubk will set the public key of the entire group;
// this is also called the master public key.
func (bns *BNGroupSigner) SetGroupPubk(groupPubk []byte) error {
	if bns == nil {
		return ErrInvalid
	}
	pubkpoint := new(bn256.G2)
	_, err := pubkpoint.Unmarshal(groupPubk)
	if err != nil {
		return err
	}
	bns.groupPubk = pubkpoint
	return nil
}

// VerifyGroupShares checks groupShares to ensure that it can be used as
// a valid ordering of the validators to correctly compute valid
// group signatures.
//
// We first check to make sure that each public key is a valid element of
// bn256.G2. From there, we also check to make sure that the byte slice did
// not appear in a previous position; to do this, we use a hash map.
func (bns *BNGroupSigner) VerifyGroupShares(groupShares [][]byte) error {
	var keyArr [constants.CurveBN256EthPubkeyLen]byte
	pubk := new(bn256.G2)
	kmap := make(map[[constants.CurveBN256EthPubkeyLen]byte]bool)
	for k := 0; k < len(groupShares); k++ {
		sliceK := groupShares[k]
		// Ensure public keys are valid
		_, err := pubk.Unmarshal(sliceK)
		if err != nil {
			return err
		}
		copy(keyArr[:], sliceK)
		// Ensure no repeated public keys
		if !kmap[keyArr] {
			kmap[keyArr] = true
		} else {
			return ErrInvalidPubkeyShares
		}
	}
	return nil
}

// PubkeyShare returns the marshalled public key of the BNGroupSigner
func (bns *BNGroupSigner) PubkeyShare() ([]byte, error) {
	if bns == nil || bns.privk == nil {
		return nil, ErrPrivkNotSet
	}
	return bns.pubk.Marshal(), nil
}

// PubkeyGroup returns the marshalled public key of the group
// (master public key).
func (bns *BNGroupSigner) PubkeyGroup() ([]byte, error) {
	if bns == nil || bns.groupPubk == nil {
		return nil, ErrPubkeyGroupNotSet
	}
	return bns.groupPubk.Marshal(), nil
}

// Sign will generate a signature for msg using the private key of the
// BNGroupSigner; this signature can be aggregated to form a valid
// group signature.
func (bns *BNGroupSigner) Sign(msg []byte) ([]byte, error) {
	if bns == nil {
		return nil, ErrInvalid
	}
	sigpoint, err := bn256.Sign(msg, bns.privk, bn256.HashToG1)
	if err != nil {
		return nil, err
	}
	return bn256.MarshalSignature(sigpoint, bns.pubk)
}

// Aggregate attempts to combine the slice of signatures in sigs into
// a group signature.
func (bns *BNGroupSigner) Aggregate(sigs [][]byte, groupShares [][]byte) ([]byte, error) {
	if bns == nil {
		return nil, ErrInvalid
	}
	err := bns.VerifyGroupShares(groupShares)
	if err != nil {
		return nil, err
	}
	if bns.groupPubk == nil {
		return nil, ErrPubkeyGroupNotSet
	}
	aggSig, err := bn256.AggregateMarshalledSignatures(sigs, groupShares, CalcThreshold(len(groupShares)))
	if err != nil {
		return nil, err
	}
	return bn256.MarshalSignature(aggSig, bns.groupPubk)
}

// BNGroupValidator is the object that performs cryptographic validation of
// BNGroupSigner signatures.
type BNGroupValidator struct {
}

// Validate will validate a BNGroupSigner signature or group signature for msg.
func (bnv *BNGroupValidator) Validate(msg []byte, sig []byte) ([]byte, error) {
	pubkey, signature, err := bn256.UnmarshalSignature(sig)
	if err != nil {
		return nil, err
	}
	ok, err := bn256.Verify(msg, signature, pubkey, bn256.HashToG1)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrInvalidSignature
	}
	pubkeyBytes := pubkey.Marshal()
	return pubkeyBytes, nil
}

// PubkeyFromSig returns the public key of the signer from the signature.
func (bnv *BNGroupValidator) PubkeyFromSig(sig []byte) ([]byte, error) {
	pubkey, _, err := bn256.UnmarshalSignature(sig)
	if err != nil {
		return nil, err
	}
	return pubkey.Marshal(), nil
}
