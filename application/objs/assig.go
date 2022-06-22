package objs

import (
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
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
func (assig *AtomicSwapSignature) UnmarshalBinary(signature []byte) error {
	if assig == nil {
		return errorz.ErrInvalid{}.New("assig.unmarshalBinary; assig not initialized")
	}
	sva, signature, err := extractSVA(signature)
	if err != nil {
		return err
	}
	assig.SVA = sva
	if err := assig.validateSVA(); err != nil {
		return err
	}
	curveSpec, signature, err := extractCurveSpec(signature)
	if err != nil {
		return err
	}
	assig.CurveSpec = curveSpec
	if err := assig.validateCurveSpec(); err != nil {
		return err
	}
	signerRole, signature, err := extractSignerRole(signature)
	if err != nil {
		return err
	}
	assig.SignerRole = signerRole
	if err := assig.validateSignerRole(); err != nil {
		return err
	}
	hashKey, signature, err := extractHash(signature)
	if err != nil {
		return err
	}
	assig.HashKey = hashKey
	signature, null, err := extractSignature(signature, curveSpec)
	if err != nil {
		return err
	}
	assig.Signature = signature
	return extractZero(null)
}

// MarshalBinary takes the AtomicSwapSignature object and returns the canonical
// byte slice
func (assig *AtomicSwapSignature) MarshalBinary() ([]byte, error) {
	if err := assig.Validate(); err != nil {
		return nil, err
	}
	signature := []byte{}
	signature = append(signature, []byte{uint8(assig.SVA)}...)
	signature = append(signature, []byte{uint8(assig.CurveSpec)}...)
	signature = append(signature, []byte{uint8(assig.SignerRole)}...)
	signature = append(signature, utils.CopySlice(assig.HashKey)...)
	signature = append(signature, utils.CopySlice(assig.Signature)...)
	return signature, nil
}

// Validate validates the AtomicSwapSignature
func (assig *AtomicSwapSignature) Validate() error {
	if assig == nil {
		return errorz.ErrInvalid{}.New("assig.validate; assig not initialized")
	}
	if err := assig.validateSVA(); err != nil {
		return err
	}
	if err := assig.validateCurveSpec(); err != nil {
		return err
	}
	if err := assig.validateSignerRole(); err != nil {
		return err
	}
	if err := utils.ValidateHash(assig.HashKey); err != nil {
		return err
	}
	return validateSignatureLen(assig.Signature, assig.CurveSpec)
}

// validateSVA validates the Signature Verification Algorithm
func (assig *AtomicSwapSignature) validateSVA() error {
	if assig == nil {
		return errorz.ErrInvalid{}.New("assig.validateSVA; assig not initialized")
	}
	if assig.SVA != HashedTimelockSVA {
		return errorz.ErrInvalid{}.New("assig.validateSVA; invalid signature verification algorithm")
	}
	return nil
}

// validateCurveSpec validates the curve specification
func (assig *AtomicSwapSignature) validateCurveSpec() error {
	if assig == nil {
		return errorz.ErrInvalid{}.New("assig.validateCurveSpec; assig not initialized")
	}
	if assig.CurveSpec != constants.CurveSecp256k1 {
		return errorz.ErrInvalid{}.New("assig.validateCurveSpec; invalid curveSpec")
	}
	return nil
}

// validateSignerRole validates the roles
func (assig *AtomicSwapSignature) validateSignerRole() error {
	if assig == nil {
		return errorz.ErrInvalid{}.New("assig.validateSignerRole; assig not initialized")
	}
	if assig.SignerRole != PrimarySignerRole && assig.SignerRole != AlternateSignerRole {
		return errorz.ErrInvalid{}.New("assig.validateSignerRole; invalid SignerRole")
	}
	return nil
}
