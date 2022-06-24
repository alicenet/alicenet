package objs

import (
	"testing"

	"github.com/alicenet/alicenet/constants"
)

func TestOwnerMarshalBinary(t *testing.T) {
	onr := &Owner{}
	_, err := onr.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	onr.CurveSpec = constants.CurveSecp256k1
	_, err = onr.MarshalBinary()
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	onr.Account = make([]byte, constants.OwnerLen)
	_, err = onr.MarshalBinary()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestOwnerUnmarshalBinary(t *testing.T) {
	onr := &Owner{}
	data := make([]byte, 0)
	err := onr.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}

	data = make([]byte, 1)
	err = onr.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}

	data = make([]byte, 1+constants.OwnerLen+1)
	err = onr.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (3)")
	}

	data = make([]byte, 1+constants.OwnerLen)
	err = onr.UnmarshalBinary(data)
	if err == nil {
		t.Fatal("Should raise an error (4)")
	}

	data = make([]byte, 1+constants.OwnerLen)
	data[0] = uint8(constants.CurveSecp256k1)
	err = onr.UnmarshalBinary(data)
	if err != nil {
		t.Fatalf("Should pass: %v", err)
	}
}

func TestOwnerNew(t *testing.T) {
	onr := &Owner{}
	curveSpec := constants.CurveSpec(0)
	acct := make([]byte, 0)
	err := onr.New(acct, curveSpec)
	if err == nil {
		t.Fatal("Should raise an error (1)")
	}
	if onr.CurveSpec != 0 {
		t.Fatal("Should revert to 0")
	}
	if onr.Account != nil {
		t.Fatal("Should revert to nil")
	}

	curveSpec = constants.CurveSecp256k1
	err = onr.New(acct, curveSpec)
	if err == nil {
		t.Fatal("Should raise an error (2)")
	}
	if onr.CurveSpec != 0 {
		t.Fatal("Should revert to 0")
	}
	if onr.Account != nil {
		t.Fatal("Should revert to nil")
	}

	curveSpec = constants.CurveSecp256k1
	acct = make([]byte, constants.OwnerLen)
	err = onr.New(acct, curveSpec)
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestOwnerNewFromASO(t *testing.T) {
	onr := &Owner{}
	aso := &AtomicSwapOwner{}
	err := onr.NewFromAtomicSwapOwner(aso)
	if err == nil {
		t.Fatal("Should raise an error")
	}

	priOwner := &Owner{}
	priOwner.CurveSpec = constants.CurveSecp256k1
	priOwner.Account = make([]byte, constants.OwnerLen)
	altOwner := &Owner{}
	altOwner.CurveSpec = constants.CurveSecp256k1
	altOwner.Account = make([]byte, constants.OwnerLen)
	hashLock := make([]byte, constants.HashLen)
	err = aso.NewFromOwner(priOwner, altOwner, hashLock)
	if err != nil {
		t.Fatal(err)
	}
	err = onr.NewFromAtomicSwapOwner(aso)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOwnerNewFromASSO(t *testing.T) {
	onr := &Owner{}
	asso := &AtomicSwapSubOwner{}
	err := onr.NewFromAtomicSwapSubOwner(asso)
	if err == nil {
		t.Fatal("Should raise an error")
	}

	asso.CurveSpec = constants.CurveSecp256k1
	asso.Account = make([]byte, constants.OwnerLen)
	err = onr.NewFromAtomicSwapSubOwner(asso)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOwnerNewFromDSO(t *testing.T) {
	onr := &Owner{}
	dso := &DataStoreOwner{}
	err := onr.NewFromDataStoreOwner(dso)
	if err == nil {
		t.Fatal("Should raise an error")
	}

	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	dso.New(acct, curveSpec)
	err = onr.NewFromDataStoreOwner(dso)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOwnerNewFromVSO(t *testing.T) {
	onr := &Owner{}
	vso := &ValueStoreOwner{}
	err := onr.NewFromValueStoreOwner(vso)
	if err == nil {
		t.Fatal("Should raise an error")
	}

	curveSpec := constants.CurveSecp256k1
	acct := make([]byte, constants.OwnerLen)
	vso.New(acct, curveSpec)
	err = onr.NewFromValueStoreOwner(vso)
	if err != nil {
		t.Fatal(err)
	}
}

func TestOwnerValidate(t *testing.T) {
	onr := &Owner{}
	err := onr.Validate()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	onr.CurveSpec = constants.CurveSecp256k1
	err = onr.Validate()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	onr.Account = make([]byte, constants.OwnerLen)
	err = onr.Validate()
	if err != nil {
		t.Fatalf("Should pass: %v", err)
	}
}
