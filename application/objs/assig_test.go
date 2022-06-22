package objs

import (
	"testing"

	"github.com/alicenet/alicenet/constants"
)

func TestASSigMarshalBinary(t *testing.T) {
	asSig := &AtomicSwapSignature{}
	_, err := asSig.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised an error (1)")
	}

	asSig.SVA = HashedTimelockSVA
	_, err = asSig.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised an error (2)")
	}

	asSig.CurveSpec = constants.CurveSecp256k1
	_, err = asSig.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised an error (3)")
	}

	asSig.SignerRole = PrimarySignerRole
	_, err = asSig.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised an error (4)")
	}

	asSig.HashKey = make([]byte, 32)
	_, err = asSig.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised an error (5)")
	}

	asSig.Signature = make([]byte, constants.CurveSecp256k1SigLen)
	_, err = asSig.MarshalBinary()
	if err != nil {
		t.Fatal("Should pass")
	}
}

func TestASSigUnmarshalBinary(t *testing.T) {
	asSig := &AtomicSwapSignature{}

	signature := make([]byte, 0)
	err := asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (1)")
	}

	signature = make([]byte, 1)
	err = asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (2)")
	}

	signature = make([]byte, 1)
	signature[0] = uint8(HashedTimelockSVA)
	err = asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (3)")
	}

	signature = make([]byte, 2)
	signature[0] = uint8(HashedTimelockSVA)
	err = asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (4)")
	}

	signature = make([]byte, 2)
	signature[0] = uint8(HashedTimelockSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	err = asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (5)")
	}

	signature = make([]byte, 3)
	signature[0] = uint8(HashedTimelockSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	err = asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (6)")
	}

	signature = make([]byte, 3)
	signature[0] = uint8(HashedTimelockSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	signature[2] = uint8(PrimarySignerRole)
	err = asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (7)")
	}

	signature = make([]byte, 3+32-1)
	signature[0] = uint8(HashedTimelockSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	signature[2] = uint8(PrimarySignerRole)
	err = asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (8)")
	}

	signature = make([]byte, 3+32)
	signature[0] = uint8(HashedTimelockSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	signature[2] = uint8(PrimarySignerRole)
	err = asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (9)")
	}

	signature = make([]byte, 3+32+constants.CurveSecp256k1SigLen+1)
	signature[0] = uint8(HashedTimelockSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	signature[2] = uint8(PrimarySignerRole)
	err = asSig.UnmarshalBinary(signature)
	if err == nil {
		t.Fatal("Should have raised an error (10)")
	}

	signature = make([]byte, 3+32+constants.CurveSecp256k1SigLen)
	signature[0] = uint8(HashedTimelockSVA)
	signature[1] = uint8(constants.CurveSecp256k1)
	signature[2] = uint8(PrimarySignerRole)
	err = asSig.UnmarshalBinary(signature)
	if err != nil {
		t.Fatal("Should pass")
	}
}
