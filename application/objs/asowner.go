package objs

import (
	"bytes"

	"github.com/alicenet/alicenet/errorz"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

// AtomicSwapOwner describes the necessary information for AtomicSwap object
type AtomicSwapOwner struct {
	SVA            SVA
	HashLock       []byte
	AlternateOwner *AtomicSwapSubOwner
	PrimaryOwner   *AtomicSwapSubOwner
}

// New creates a new AtomicSwapOwner
func (aso *AtomicSwapOwner) New(priOwnerAcct []byte, altOwnerAcct []byte, hashKey []byte) error {
	if aso == nil {
		return errorz.ErrInvalid{}.New("aso.new; aso not initialized")
	}
	if len(hashKey) != constants.HashLen {
		return errorz.ErrInvalid{}.New("aso.new; invalid hashKey")
	}
	if len(priOwnerAcct) != constants.OwnerLen {
		return errorz.ErrInvalid{}.New("aso.new; invalid primary account length")
	}
	if len(altOwnerAcct) != constants.OwnerLen {
		return errorz.ErrInvalid{}.New("aso.new; invalid alternate account length")
	}
	aso.SVA = HashedTimelockSVA
	aso.HashLock = crypto.Hasher(hashKey)
	aso.PrimaryOwner = &AtomicSwapSubOwner{
		CurveSpec: constants.CurveSecp256k1,
		Account:   utils.CopySlice(priOwnerAcct),
	}
	aso.AlternateOwner = &AtomicSwapSubOwner{
		CurveSpec: constants.CurveSecp256k1,
		Account:   utils.CopySlice(altOwnerAcct),
	}
	return nil
}

// NewFromOwner creates a new AtomicSwapOwner from Owner objects
func (aso *AtomicSwapOwner) NewFromOwner(priOwner *Owner, altOwner *Owner, hashKey []byte) error {
	if aso == nil {
		return errorz.ErrInvalid{}.New("aso.newFromOwner; aso not initialized")
	}
	if len(hashKey) != constants.HashLen {
		return errorz.ErrInvalid{}.New("aso.newFromOwner; invalid hashKey")
	}
	aso.SVA = HashedTimelockSVA
	aso.HashLock = crypto.Hasher(hashKey)
	aso.PrimaryOwner = &AtomicSwapSubOwner{}
	err := aso.PrimaryOwner.NewFromOwner(priOwner)
	if err != nil {
		aso.SVA = 0
		aso.HashLock = nil
		aso.PrimaryOwner = nil
		aso.AlternateOwner = nil
		return err
	}
	aso.AlternateOwner = &AtomicSwapSubOwner{}
	err = aso.AlternateOwner.NewFromOwner(altOwner)
	if err != nil {
		aso.SVA = 0
		aso.HashLock = nil
		aso.PrimaryOwner = nil
		aso.AlternateOwner = nil
		return err
	}
	return nil
}

// MarshalBinary takes the AtomicSwapOwner object and returns the canonical
// byte slice
func (aso *AtomicSwapOwner) MarshalBinary() ([]byte, error) {
	if err := aso.Validate(); err != nil {
		return nil, err
	}
	owner := []byte{}
	owner = append(owner, []byte{uint8(aso.SVA)}...)
	owner = append(owner, utils.CopySlice(aso.HashLock)...)
	priOwner, err := aso.PrimaryOwner.MarshalBinary()
	if err != nil {
		return nil, err
	}
	owner = append(owner, priOwner...)
	altOwner, err := aso.AlternateOwner.MarshalBinary()
	if err != nil {
		return nil, err
	}
	owner = append(owner, altOwner...)
	return owner, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// AtomicSwapOwner object
func (aso *AtomicSwapOwner) UnmarshalBinary(o []byte) error {
	if aso == nil {
		return errorz.ErrInvalid{}.New("aso.unmarshalBinary; aso not initialized")
	}
	owner := utils.CopySlice(o)
	sva, owner, err := extractSVA(owner)
	if err != nil {
		return err
	}
	aso.SVA = sva
	if err := aso.validateSVA(); err != nil {
		return err
	}
	hashLock, owner, err := extractHash(owner)
	if err != nil {
		return err
	}
	aso.HashLock = hashLock
	priOwner := &AtomicSwapSubOwner{}
	owner, err = priOwner.UnmarshalBinary(owner)
	if err != nil {
		return err
	}
	aso.PrimaryOwner = priOwner
	altOwner := &AtomicSwapSubOwner{}
	owner, err = altOwner.UnmarshalBinary(owner)
	if err != nil {
		return err
	}
	aso.AlternateOwner = altOwner
	if err := extractZero(owner); err != nil {
		return err
	}
	if err := aso.Validate(); err != nil {
		return err
	}
	return nil
}

// PrimaryAccount returns the account byte slice of the PrimaryOwner
func (aso *AtomicSwapOwner) PrimaryAccount() ([]byte, error) {
	if aso == nil {
		return nil, errorz.ErrInvalid{}.New("aso.primaryAccount; aso not initialized")
	}
	if err := aso.PrimaryOwner.Validate(); err != nil {
		return nil, err
	}
	return utils.CopySlice(aso.PrimaryOwner.Account), nil
}

// AlternateAccount returns the account byte slice of the AlternateOwner
func (aso *AtomicSwapOwner) AlternateAccount() ([]byte, error) {
	if aso == nil {
		return nil, errorz.ErrInvalid{}.New("aso.alternateAccount; aso not initialized")
	}
	if err := aso.AlternateOwner.Validate(); err != nil {
		return nil, err
	}
	return utils.CopySlice(aso.AlternateOwner.Account), nil
}

// Validate validates the AtomicSwapOwner
func (aso *AtomicSwapOwner) Validate() error {
	if aso == nil {
		return errorz.ErrInvalid{}.New("aso.validate; aso not initialized")
	}
	if err := aso.validateSVA(); err != nil {
		return err
	}
	if err := utils.ValidateHash(aso.HashLock); err != nil {
		return err
	}
	if err := aso.AlternateOwner.Validate(); err != nil {
		return err
	}
	if err := aso.PrimaryOwner.Validate(); err != nil {
		return err
	}
	return nil
}

// ValidateSignature validates the signature
func (aso *AtomicSwapOwner) ValidateSignature(msg []byte, sig *AtomicSwapSignature, isExpired bool) error {
	if err := aso.Validate(); err != nil {
		return errorz.ErrInvalid{}.New("aso.validateSignature; invalid AtomicSwapOwner")
	}
	if err := sig.Validate(); err != nil {
		return errorz.ErrInvalid{}.New("aso.validateSignature; invalid AtomicSwapSignature")
	}
	if aso.SVA != sig.SVA {
		return errorz.ErrInvalid{}.New("aso.validateSignature; incorrect SVA")
	}
	hsh := crypto.Hasher(sig.HashKey)
	if !bytes.Equal(aso.HashLock, hsh) {
		return errorz.ErrInvalid{}.New("aso.validateSignature; incorrect hash key")
	}
	switch sig.SignerRole {
	case PrimarySignerRole:
		if !isExpired {
			return errorz.ErrInvalid{}.New("aso.validateSignature; PrimaryOwner can not sign before expiration")
		}
		if err := aso.PrimaryOwner.ValidateSignature(msg, sig); err != nil {
			return err
		}
		return nil
	case AlternateSignerRole:
		if isExpired {
			return errorz.ErrInvalid{}.New("aso.validateSignature; AlternateOwner can not sign after expiration")
		}
		if err := aso.AlternateOwner.ValidateSignature(msg, sig); err != nil {
			return err
		}
		return nil
	default:
		return errorz.ErrInvalid{}.New("aso.validateSignature; invalid signerRole")
	}
}

// validateSVA validates the Signature Verification Algorithm
func (aso *AtomicSwapOwner) validateSVA() error {
	if aso == nil {
		return errorz.ErrInvalid{}.New("aso.validateSVA; aso not initialized")
	}
	if aso.SVA != HashedTimelockSVA {
		return errorz.ErrInvalid{}.New("aso.validateSVA; invalid signature verification algorithm")
	}
	return nil
}

// SignAsPrimary ...
func (aso *AtomicSwapOwner) SignAsPrimary(msg []byte, signer *crypto.Secp256k1Signer, hashKey []byte) (*AtomicSwapSignature, error) {
	if aso == nil {
		return nil, errorz.ErrInvalid{}.New("aso.signAsPrimary; aso not initialized")
	}
	sig, err := signer.Sign(msg)
	if err != nil {
		return nil, err
	}
	hsh := crypto.Hasher(hashKey)
	if !bytes.Equal(hsh, aso.HashLock) {
		return nil, errorz.ErrInvalid{}.New("aso.signAsPrimary; invalid hash key")
	}
	s := &AtomicSwapSignature{
		SVA:        HashedTimelockSVA,
		CurveSpec:  constants.CurveSecp256k1,
		SignerRole: PrimarySignerRole,
		HashKey:    hashKey,
		Signature:  sig,
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return s, nil
}

// SignAsAlternate ...
func (aso *AtomicSwapOwner) SignAsAlternate(msg []byte, signer *crypto.Secp256k1Signer, hashKey []byte) (*AtomicSwapSignature, error) {
	if aso == nil {
		return nil, errorz.ErrInvalid{}.New("aso.signAsAlternate; aso not initialized")
	}
	sig, err := signer.Sign(msg)
	if err != nil {
		return nil, err
	}
	hsh := crypto.Hasher(hashKey)
	if !bytes.Equal(hsh, aso.HashLock) {
		return nil, errorz.ErrInvalid{}.New("aso.signAsAlternate; invalid hash key")
	}
	s := &AtomicSwapSignature{
		SVA:        HashedTimelockSVA,
		CurveSpec:  constants.CurveSecp256k1,
		SignerRole: AlternateSignerRole,
		HashKey:    hashKey,
		Signature:  sig,
	}
	if err := s.Validate(); err != nil {
		return nil, err
	}
	return s, nil
}

// AtomicSwapSubOwner ...
type AtomicSwapSubOwner struct {
	CurveSpec constants.CurveSpec
	Account   []byte
}

// NewFromOwner takes an Owner object and creates the corresponding
// AtomicSwapSubOwner
func (asso *AtomicSwapSubOwner) NewFromOwner(o *Owner) error {
	if asso == nil {
		return errorz.ErrInvalid{}.New("asso.newFromOwner; asso not initialized")
	}
	if err := o.Validate(); err != nil {
		return err
	}
	asso.CurveSpec = o.CurveSpec
	asso.Account = utils.CopySlice(o.Account)
	if err := asso.Validate(); err != nil {
		asso.CurveSpec = 0
		asso.Account = nil
		return err
	}
	return nil
}

// MarshalBinary takes the AtomicSwapSubOwner object and returns the canonical
// byte slice
func (asso *AtomicSwapSubOwner) MarshalBinary() ([]byte, error) {
	if err := asso.Validate(); err != nil {
		return nil, err
	}
	var owner []byte
	owner = append(owner, []byte{uint8(asso.CurveSpec)}...)
	owner = append(owner, utils.CopySlice(asso.Account)...)
	return owner, nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// AtomicSwapSubOwner object
func (asso *AtomicSwapSubOwner) UnmarshalBinary(o []byte) ([]byte, error) {
	owner := utils.CopySlice(o)
	curveSpec, owner, err := extractCurveSpec(owner)
	if err != nil {
		return nil, err
	}
	account, owner, err := extractAccount(owner)
	if err != nil {
		return nil, err
	}
	asso.CurveSpec = curveSpec
	asso.Account = account
	if err := asso.Validate(); err != nil {
		return nil, err
	}
	return owner, nil
}

// ValidateSignature validates the signature of the AtomicSwapSignature object
func (asso *AtomicSwapSubOwner) ValidateSignature(msg []byte, sig *AtomicSwapSignature) error {
	if err := asso.Validate(); err != nil {
		return errorz.ErrInvalid{}.New("asso.validateSignature; invalid AtomicSwapSubOwner")
	}
	if asso.CurveSpec != sig.CurveSpec {
		return errorz.ErrInvalid{}.New("asso.validateSignature; mismatched curve spec")
	}
	val := crypto.Secp256k1Validator{}
	pk, err := val.Validate(msg, sig.Signature)
	if err != nil {
		return err
	}
	account := crypto.GetAccount(pk)
	if !bytes.Equal(account, asso.Account) {
		return errorz.ErrInvalid{}.New("asso.validateSignature; invalid sig for account")
	}
	return nil
}

// validateCurveSpec validates the curve specification for AtomicSwapSubOwner
func (asso *AtomicSwapSubOwner) validateCurveSpec() error {
	if asso.CurveSpec != constants.CurveSecp256k1 {
		return errorz.ErrInvalid{}.New("asso.validateCurveSpec; invalid curveSpec")
	}
	return nil
}

// validateAccount validates the account for AtomicSwapSubOwner
func (asso *AtomicSwapSubOwner) validateAccount() error {
	if len(asso.Account) != constants.OwnerLen {
		return errorz.ErrInvalid{}.New("asso.validateAccount; incorrect account length")
	}
	return nil
}

// Validate validates the AtomicSwapSubOwner object
func (asso *AtomicSwapSubOwner) Validate() error {
	if asso == nil {
		return errorz.ErrInvalid{}.New("asso.validate; asso not initialized")
	}
	if err := asso.validateCurveSpec(); err != nil {
		return err
	}
	if err := asso.validateAccount(); err != nil {
		return err
	}
	return nil
}
