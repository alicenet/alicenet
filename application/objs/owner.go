package objs

import (
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"
)

// Owner contains information related to a general owner object
type Owner struct {
	CurveSpec constants.CurveSpec
	Account   []byte
}

// New makes a new Owner
func (onr *Owner) New(acct []byte, curveSpec constants.CurveSpec) error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("owner.new; owner not initialized")
	}
	onr.CurveSpec = curveSpec
	onr.Account = acct
	if err := onr.Validate(); err != nil {
		onr.CurveSpec = 0
		onr.Account = nil
		return err
	}
	return nil
}

// NewFromAtomicSwapOwner makes a new Owner from an AtomicSwapOwner
// (PrimaryOwner)
func (onr *Owner) NewFromAtomicSwapOwner(aso *AtomicSwapOwner) error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("owner.newFromAtomicSwapOwner; owner not initialized")
	}
	if err := aso.Validate(); err != nil {
		return err
	}
	return onr.New(aso.PrimaryOwner.Account, aso.PrimaryOwner.CurveSpec)
}

// NewFromAtomicSwapSubOwner makes a new Owner from an AtomicSwapSubOwner
func (onr *Owner) NewFromAtomicSwapSubOwner(asso *AtomicSwapSubOwner) error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("owner.newFromAtomicSwapSubOwner; owner not initialized")
	}
	if err := asso.Validate(); err != nil {
		return err
	}
	return onr.New(asso.Account, asso.CurveSpec)
}

// NewFromDataStoreOwner makes a new Owner from a DataStoreOwner
func (onr *Owner) NewFromDataStoreOwner(dso *DataStoreOwner) error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("owner.newFromDataStoreOwner; owner not initialized")
	}
	if err := dso.Validate(); err != nil {
		return err
	}
	return onr.New(dso.Account, dso.CurveSpec)
}

// NewFromValueStoreOwner makes a new Owner from a ValueStoreOwner
func (onr *Owner) NewFromValueStoreOwner(vso *ValueStoreOwner) error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("owner.newFromValueStoreOwner; owner not initialized")
	}
	if err := vso.Validate(); err != nil {
		return err
	}
	return onr.New(vso.Account, vso.CurveSpec)
}

// MarshalBinary takes the Owner object and returns the canonical
// byte slice
func (onr *Owner) MarshalBinary() ([]byte, error) {
	if err := onr.Validate(); err != nil {
		return nil, err
	}
	owner := []byte{}
	owner = append(owner, []byte{uint8(onr.CurveSpec)}...)
	owner = append(owner, utils.CopySlice(onr.Account)...)
	return owner, nil
}

// Validate validates the Owner object
func (onr *Owner) Validate() error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("owner.validate; owner not initialized")
	}
	if err := onr.validateCurveSpec(); err != nil {
		return err
	}
	if err := onr.validateAccount(); err != nil {
		return err
	}
	return nil
}

// UnmarshalBinary takes a byte slice and returns the corresponding
// Owner object
func (onr *Owner) UnmarshalBinary(o []byte) error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("owner.unmarshalBinary; owner not initialized")
	}
	owner := utils.CopySlice(o)
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
	onr.CurveSpec = curveSpec
	onr.Account = account
	if err := onr.Validate(); err != nil {
		return err
	}
	return nil
}

func (onr *Owner) validateCurveSpec() error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("owner.validateCurveSpec; owner not initialized")
	}
	if onr.CurveSpec == 0 {
		return errorz.ErrInvalid{}.New("owner.validateCurveSpec; invalid curve spec")
	}
	return nil
}

func (onr *Owner) validateAccount() error {
	if onr == nil {
		return errorz.ErrInvalid{}.New("owner.validateAccount; owner not initialized")
	}
	if len(onr.Account) != constants.OwnerLen {
		return errorz.ErrInvalid{}.New("owner.validateAccount; invalid account length")
	}
	return nil
}
