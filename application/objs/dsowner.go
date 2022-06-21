package objs

import (
	"bytes"

	"github.com/alicenet/alicenet/errorz"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/utils"
)

// DataStoreOwner is the struct which specifies the owner of the DataStore
type DataStoreOwner struct {
	SVA       SVA
	CurveSpec constants.CurveSpec
	Account   []byte
}

// New makes a new DataStoreOwner
func (dso *DataStoreOwner) New(acct []byte, curveSpec constants.CurveSpec) {
	dso.SVA = DataStoreSVA
	dso.CurveSpec = curveSpec
	dso.Account = utils.CopySlice(acct)
}

// NewFromOwner takes an Owner object and creates the corresponding
// DataStoreOwner
func (dso *DataStoreOwner) NewFromOwner(o *Owner) error {
	if dso == nil {
		return errorz.ErrInvalid{}.New("dso.newFromOwner; dso not initialized")
	}
	if err := o.Validate(); err != nil {
		return err
	}
	dso.New(o.Account, o.CurveSpec)
	if err := dso.Validate(); err != nil {
		dso.SVA = 0
		dso.CurveSpec = 0
		dso.Account = nil
		return err
	}
	return nil
}

// MarshalBinary takes the DataStoreOwner object and returns the canonical
// byte slice
func (dso *DataStoreOwner) MarshalBinary() ([]byte, error) {
	if err := dso.Validate(); err != nil {
		return nil, err
	}
	owner := []byte{}
	owner = append(owner, []byte{uint8(dso.SVA)}...)
	owner = append(owner, []byte{uint8(dso.CurveSpec)}...)
	owner = append(owner, utils.CopySlice(dso.Account)...)
	return owner, nil
}

// Validate validates the DataStoreOwner
func (dso *DataStoreOwner) Validate() error {
	if dso == nil {
		return errorz.ErrInvalid{}.New("dso.validate; dso not initialized")
	}
	if err := dso.validateSVA(); err != nil {
		return err
	}
	if err := dso.validateCurveSpec(); err != nil {
		return err
	}
	if err := dso.validateAccount(); err != nil {
		return err
	}
	return nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// DataStoreOwner object
func (dso *DataStoreOwner) UnmarshalBinary(o []byte) error {
	if dso == nil {
		return errorz.ErrInvalid{}.New("dso.unmarshalBinary; dso not initialized")
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
	dso.SVA = sva
	dso.CurveSpec = curveSpec
	dso.Account = account
	if err := dso.Validate(); err != nil {
		return err
	}
	return nil
}

// ValidateSignature validates the DataStoreSignature
func (dso *DataStoreOwner) ValidateSignature(msg []byte, sig *DataStoreSignature, isExpired bool) error {
	if err := dso.Validate(); err != nil {
		return errorz.ErrInvalid{}.New("dso.validateSignature; invalid DataStoreOwner")
	}
	if err := sig.Validate(); err != nil {
		return errorz.ErrInvalid{}.New("dso.validateSignature; invalid DataStoreSignature")
	}
	if !isExpired && sig.CurveSpec != dso.CurveSpec {
		return errorz.ErrInvalid{}.New("dso.validateSignature; unmatched curve spec")
	}
	switch sig.CurveSpec {
	case constants.CurveSecp256k1:
		val := crypto.Secp256k1Validator{}
		pk, err := val.Validate(msg, sig.Signature)
		if err != nil {
			return err
		}
		if !isExpired {
			account := crypto.GetAccount(pk)
			if !bytes.Equal(account, dso.Account) {
				return errorz.ErrInvalid{}.New("dso.validateSignature; invalid sig for secp256k1 account")
			}
		}
		return nil
	case constants.CurveBN256Eth:
		val := crypto.BNValidator{}
		pk, err := val.Validate(msg, sig.Signature)
		if err != nil {
			return err
		}
		if !isExpired {
			account := crypto.GetAccount(pk)
			if !bytes.Equal(account, dso.Account) {
				return errorz.ErrInvalid{}.New("dso.validateSignature; invalid sig for bn256 account")
			}
		}
		return nil
	default:
		return errorz.ErrInvalid{}.New("dso.validateSignature; invalid curve spec")
	}
}

func (dso *DataStoreOwner) validateCurveSpec() error {
	if dso == nil {
		return errorz.ErrInvalid{}.New("dso.validateCurveSpec; dso not initialized")
	}
	if !(dso.CurveSpec == constants.CurveSecp256k1) && !(dso.CurveSpec == constants.CurveBN256Eth) {
		return errorz.ErrInvalid{}.New("dso.validateCurveSpec; invalid curve spec")
	}
	return nil
}

func (dso *DataStoreOwner) validateSVA() error {
	if dso == nil {
		return errorz.ErrInvalid{}.New("dso.validateSVA; dso not initialized")
	}
	if dso.SVA != DataStoreSVA {
		return errorz.ErrInvalid{}.New("dso.validateSVA; invalid signature verification algorithm")
	}
	return nil
}

func (dso *DataStoreOwner) validateAccount() error {
	if dso == nil {
		return errorz.ErrInvalid{}.New("dso.validateAccount; dso not initialized")
	}
	if len(dso.Account) != constants.OwnerLen {
		return errorz.ErrInvalid{}.New("dso.validateAccount; dso.account has incorrect length")
	}
	return nil
}

// Sign allows has the DataStoreOwner sign the message msg with signer s
func (dso *DataStoreOwner) Sign(msg []byte, s Signer) (*DataStoreSignature, error) {
	sig := &DataStoreSignature{
		SVA: DataStoreSVA,
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
		return nil, errorz.ErrInvalid{}.New("dso.sign; invalid signer type")
	}
}

// DataStoreSignature ...
type DataStoreSignature struct {
	SVA       SVA
	CurveSpec constants.CurveSpec
	Signature []byte
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// DataStoreSignature object
func (dss *DataStoreSignature) UnmarshalBinary(signature []byte) error {
	if dss == nil {
		return errorz.ErrInvalid{}.New("dss.unmarshalBinary; dss not initialized")
	}
	sva, signature, err := extractSVA(signature)
	if err != nil {
		return err
	}
	dss.SVA = sva
	curveSpec, signature, err := extractCurveSpec(signature)
	if err != nil {
		return err
	}
	dss.CurveSpec = curveSpec
	signature, null, err := extractSignature(signature, curveSpec)
	if err != nil {
		return err
	}
	dss.Signature = signature
	if err := extractZero(null); err != nil {
		return err
	}
	return dss.Validate()
}

// MarshalBinary takes the DataStoreSignature object and returns the canonical
// byte slice
func (dss *DataStoreSignature) MarshalBinary() ([]byte, error) {
	if err := dss.Validate(); err != nil {
		return nil, err
	}
	signature := []byte{}
	signature = append(signature, []byte{uint8(dss.SVA)}...)
	signature = append(signature, []byte{uint8(dss.CurveSpec)}...)
	signature = append(signature, utils.CopySlice(dss.Signature)...)
	return signature, nil
}

// Validate validates the DataStoreSignature object
func (dss *DataStoreSignature) Validate() error {
	if dss == nil {
		return errorz.ErrInvalid{}.New("dss.validate; dss not initialized")
	}
	if err := dss.validateSVA(); err != nil {
		return err
	}
	if err := dss.validateCurveSpec(); err != nil {
		return err
	}
	return validateSignatureLen(dss.Signature, dss.CurveSpec)
}

func (dss *DataStoreSignature) validateSVA() error {
	if dss == nil {
		return errorz.ErrInvalid{}.New("dss.validateSVA; dss not initialized")
	}
	if dss.SVA != DataStoreSVA {
		return errorz.ErrInvalid{}.New("dss.validateSVA; invalid signature verification algorithm")
	}
	return nil
}

func (dss *DataStoreSignature) validateCurveSpec() error {
	if dss == nil {
		return errorz.ErrInvalid{}.New("dss.validateCurveSpec; dss not initialized")
	}
	if !(dss.CurveSpec == constants.CurveSecp256k1) && !(dss.CurveSpec == constants.CurveBN256Eth) {
		return errorz.ErrInvalid{}.New("dss.validateCurveSpec; invalid curve spec")
	}
	return nil
}
