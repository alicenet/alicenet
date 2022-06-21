package objs

import (
	"bytes"

	"github.com/alicenet/alicenet/errorz"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

// ValueStoreOwner contains information related to the owner of the ValueStore
type ValueStoreOwner struct {
	SVA       SVA
	CurveSpec constants.CurveSpec
	Account   []byte
}

// New makes a new ValueStoreOwner
func (vso *ValueStoreOwner) New(acct []byte, curveSpec constants.CurveSpec) {
	vso.SVA = ValueStoreSVA
	vso.CurveSpec = curveSpec
	vso.Account = utils.CopySlice(acct)
}

// NewFromOwner takes an Owner object and creates the corresponding
// ValueStoreOwner
func (vso *ValueStoreOwner) NewFromOwner(o *Owner) error {
	if vso == nil {
		return errorz.ErrInvalid{}.New("vso.newFromOwner; vso not initialized")
	}
	if err := o.Validate(); err != nil {
		return err
	}
	vso.New(o.Account, o.CurveSpec)
	if err := vso.Validate(); err != nil {
		vso.SVA = 0
		vso.CurveSpec = 0
		vso.Account = nil
		return err
	}
	return nil
}

// MarshalBinary takes the ValueStoreOwner object and returns the canonical
// byte slice
func (vso *ValueStoreOwner) MarshalBinary() ([]byte, error) {
	if err := vso.Validate(); err != nil {
		return nil, err
	}
	owner := []byte{}
	owner = append(owner, []byte{uint8(vso.SVA)}...)
	owner = append(owner, []byte{uint8(vso.CurveSpec)}...)
	owner = append(owner, utils.CopySlice(vso.Account)...)
	return owner, nil
}

// Validate validates the ValueStoreOwner object
func (vso *ValueStoreOwner) Validate() error {
	if vso == nil {
		return errorz.ErrInvalid{}.New("vso.validate; vso not initialized")
	}
	if err := vso.validateSVA(); err != nil {
		return err
	}
	if err := vso.validateCurveSpec(); err != nil {
		return err
	}
	if err := vso.validateAccount(); err != nil {
		return err
	}
	return nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// ValueStoreOwner object
func (vso *ValueStoreOwner) UnmarshalBinary(o []byte) error {
	if vso == nil {
		return errorz.ErrInvalid{}.New("vso.unmarshalBinary; vso not initialized")
	}
	owner := utils.CopySlice(o)
	sva, owner, err := extractSVA(owner)
	if err != nil {
		return err
	}
	curveSpec, owner, err := extractCurveSpec(owner)
	if err != nil {
		return err
	}
	account, owner, err := extractAccount(owner)
	if err != nil {
		return err
	}
	if err := extractZero(owner); err != nil {
		return err
	}
	vso.SVA = sva
	vso.CurveSpec = curveSpec
	vso.Account = account
	if err := vso.Validate(); err != nil {
		return err
	}
	return nil
}

// ValidateSignature validates ValueStoreSignature sig for message msg
func (vso *ValueStoreOwner) ValidateSignature(msg []byte, sig *ValueStoreSignature) error {
	if err := vso.Validate(); err != nil {
		return errorz.ErrInvalid{}.New("vso.validateSignature; invalid ValueStoreOwner")
	}
	if err := sig.Validate(); err != nil {
		return errorz.ErrInvalid{}.New("vso.validateSignature; invalid ValueStoreSignature")
	}
	if vso.CurveSpec != sig.CurveSpec {
		return errorz.ErrInvalid{}.New("vso.validateSignature; mismatched curve spec")
	}
	signature := sig.Signature
	switch vso.CurveSpec {
	case constants.CurveSecp256k1:
		val := crypto.Secp256k1Validator{}
		pk, err := val.Validate(msg, signature)
		if err != nil {
			return err
		}
		account := crypto.GetAccount(pk)
		if !bytes.Equal(account, vso.Account) {
			return errorz.ErrInvalid{}.New("vso.validateSignature; invalid sig for secp256k1 account")
		}
		return nil
	case constants.CurveBN256Eth:
		val := crypto.BNValidator{}
		pk, err := val.Validate(msg, signature)
		if err != nil {
			return err
		}
		account := crypto.GetAccount(pk)
		if !bytes.Equal(account, vso.Account) {
			return errorz.ErrInvalid{}.New("vso.validateSignature; invalid sig for bn256 account")
		}
		return nil
	default:
		return errorz.ErrInvalid{}.New("vso.validateSignature; invalid curve spec")
	}
}

func (vso *ValueStoreOwner) validateCurveSpec() error {
	if vso == nil {
		return errorz.ErrInvalid{}.New("vso.validateCurveSpec; vso not initialized")
	}
	if !(vso.CurveSpec == constants.CurveSecp256k1) && !(vso.CurveSpec == constants.CurveBN256Eth) {
		return errorz.ErrInvalid{}.New("vso.validateCurveSpec; invalid curve spec")
	}
	return nil
}

func (vso *ValueStoreOwner) validateSVA() error {
	if vso == nil {
		return errorz.ErrInvalid{}.New("vso.validateSVA; vso not initialized")
	}
	if vso.SVA != ValueStoreSVA {
		return errorz.ErrInvalid{}.New("vso.validateSVA; invalid signature verification algorithm")
	}
	return nil
}

func (vso *ValueStoreOwner) validateAccount() error {
	if vso == nil {
		return errorz.ErrInvalid{}.New("vso.validateAccount; vso not initialized")
	}
	if len(vso.Account) != constants.OwnerLen {
		return errorz.ErrInvalid{}.New("vso.validateAccount; vso.account has incorrect length")
	}
	return nil
}

// Sign signs message msg with signer s
func (vso *ValueStoreOwner) Sign(msg []byte, s Signer) (*ValueStoreSignature, error) {
	sig := &ValueStoreSignature{
		SVA: ValueStoreSVA,
	}
	switch s.(type) {
	case *crypto.Secp256k1Signer:
		sig.CurveSpec = constants.CurveSecp256k1
		signature, err := s.Sign(msg)
		if err != nil {
			return nil, err
		}
		sig.Signature = signature
		return sig, nil
	case *crypto.BNSigner:
		sig.CurveSpec = constants.CurveBN256Eth
		signature, err := s.Sign(msg)
		if err != nil {
			return nil, err
		}
		sig.Signature = signature
		return sig, nil
	default:
		return nil, errorz.ErrInvalid{}.New("vso.sign; invalid signer type")
	}
}

// ValueStoreSignature is a struct which the necessary information
// for signing a ValueStore
type ValueStoreSignature struct {
	SVA       SVA
	CurveSpec constants.CurveSpec
	Signature []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// ValueStoreSignature object
func (vss *ValueStoreSignature) UnmarshalBinary(signature []byte) error {
	if vss == nil {
		return errorz.ErrInvalid{}.New("vss.unmarshalBinary; vss not initialized")
	}
	sva, signature, err := extractSVA(signature)
	if err != nil {
		return err
	}
	vss.SVA = sva
	curveSpec, signature, err := extractCurveSpec(signature)
	if err != nil {
		return err
	}
	vss.CurveSpec = curveSpec
	signature, null, err := extractSignature(signature, curveSpec)
	if err != nil {
		return err
	}
	vss.Signature = signature
	if err := extractZero(null); err != nil {
		return err
	}
	return vss.Validate()
}

// MarshalBinary takes the ValueStoreSignature object and returns the canonical
// byte slice
func (vss *ValueStoreSignature) MarshalBinary() ([]byte, error) {
	if err := vss.Validate(); err != nil {
		return nil, err
	}
	signature := []byte{}
	signature = append(signature, []byte{uint8(vss.SVA)}...)
	signature = append(signature, []byte{uint8(vss.CurveSpec)}...)
	signature = append(signature, utils.CopySlice(vss.Signature)...)
	return signature, nil
}

// Validate validates the ValueStoreSignature object
func (vss *ValueStoreSignature) Validate() error {
	if vss == nil {
		return errorz.ErrInvalid{}.New("vss.validate; vss not initialized")
	}
	if err := vss.validateSVA(); err != nil {
		return err
	}
	if err := vss.validateCurveSpec(); err != nil {
		return err
	}
	return validateSignatureLen(vss.Signature, vss.CurveSpec)
}

func (vss *ValueStoreSignature) validateSVA() error {
	if vss == nil {
		return errorz.ErrInvalid{}.New("vss.validateSVA; vss not initialized")
	}
	if vss.SVA != ValueStoreSVA {
		return errorz.ErrInvalid{}.New("vss.validateSVA; invalid signature verification algorithm")
	}
	return nil
}

func (vss *ValueStoreSignature) validateCurveSpec() error {
	if vss == nil {
		return errorz.ErrInvalid{}.New("vss.validateCurveSpec; vss not initialized")
	}
	if !(vss.CurveSpec == constants.CurveSecp256k1) && !(vss.CurveSpec == constants.CurveBN256Eth) {
		return errorz.ErrInvalid{}.New("vss.validateCurveSpec; invalid curve spec")
	}
	return nil
}
