package objs

import (
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/errorz"
	"github.com/MadBase/MadNet/utils"
)

// AtomicSwapSignature is the signs the AtomicSwap object
type AtomicSwapSignature struct {
	SVA        SVA
	CurveSpec  constants.CurveSpec
	SignerRole SignerRole
	HashKey    []byte
	Signature  []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// AtomicSwapSignature object
func (onr *AtomicSwapSignature) UnmarshalBinary(signature []byte) error {
	sva, signature, err := extractSVA(signature)
	if err != nil {
		return err
	}
	onr.SVA = sva
	if err := onr.validateSVA(); err != nil {
		return err
	}
	curveSpec, signature, err := extractCurveSpec(signature)
	if err != nil {
		return err
	}
	onr.CurveSpec = curveSpec
	if err := onr.validateCurveSpec(); err != nil {
		return err
	}
	signerRole, signature, err := extractSignerRole(signature)
	if err != nil {
		return err
	}
	onr.SignerRole = signerRole
	if err := onr.validateSignerRole(); err != nil {
		return err
	}
	hashKey, signature, err := extractHash(signature)
	if err != nil {
		return err
	}
	onr.HashKey = hashKey
	signature, null, err := extractSignature(signature, curveSpec)
	if err != nil {
		return err
	}
	onr.Signature = signature
	return extractZero(null)
}

// MarshalBinary takes the AtomicSwapSignature object and returns the canonical
// byte slice
func (onr *AtomicSwapSignature) MarshalBinary() ([]byte, error) {
	if err := onr.Validate(); err != nil {
		return nil, err
	}
	signature := []byte{}
	signature = append(signature, []byte{uint8(onr.SVA)}...)
	signature = append(signature, []byte{uint8(onr.CurveSpec)}...)
	signature = append(signature, []byte{uint8(onr.SignerRole)}...)
	signature = append(signature, utils.CopySlice(onr.HashKey)...)
	signature = append(signature, utils.CopySlice(onr.Signature)...)
	return signature, nil
}

// Validate validates the AtomicSwapSignature
func (onr *AtomicSwapSignature) Validate() error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("object is nil")
	}
	if err := onr.validateSVA(); err != nil {
		return err
	}
	if err := onr.validateCurveSpec(); err != nil {
		return err
	}
	if err := onr.validateSignerRole(); err != nil {
		return err
	}
	if err := utils.ValidateHash(onr.HashKey); err != nil {
		return err
	}
	return validateSignatureLen(onr.Signature, onr.CurveSpec)
}

// validateSVA validates the Signature Verification Algorithm
func (onr *AtomicSwapSignature) validateSVA() error {
	if onr.SVA != HashedTimelockSVA {
		return errorz.ErrInvalid{}.New("signature verification algorithm invalid for AtomicSwapSignature")
	}
	return nil
}

// validateCurveSpec validates the curve specification
func (onr *AtomicSwapSignature) validateCurveSpec() error {
	if onr.CurveSpec != constants.CurveSecp256k1 {
		return errorz.ErrInvalid{}.New("Invalid curveSpec for AtomicSwapSignature")
	}
	return nil
}

// validateSignerRole validates the roles
func (onr *AtomicSwapSignature) validateSignerRole() error {
	if onr.SignerRole != PrimarySignerRole && onr.SignerRole != AlternateSignerRole {
		return errorz.ErrInvalid{}.New("Invalid SignerRole for AtomicSwapSignature")
	}
	return nil
}
