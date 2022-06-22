package objs

import (
	"bytes"
	"testing"

	"github.com/alicenet/alicenet/application/objs/uint256"
)

func getTestingStruct() PaginationToken {
	utxoid := make([]byte, 32)
	for i := range utxoid {
		utxoid[i] = byte(i)
	}

	return PaginationToken{LastPaginatedType: LastPaginatedDeposit, TotalValue: uint256.Two(), LastKey: utxoid}
}

func TestUnMarshalInvalid(t *testing.T) {
	p := getTestingStruct()
	if err := p.UnmarshalBinary(nil); err == nil {
		t.Fatal("Should raise an error when called with nil byte slice")
	}

	if err := p.UnmarshalBinary(make([]byte, 10)); err == nil {
		t.Fatal("Should raise an error when called with byte slice of incorrect size")
	}

	b := make([]byte, 65)
	b[0] = 2

	if err := p.UnmarshalBinary(b); err == nil {
		t.Fatal("Should raise an error when called with invalid LastPaginatedType")
	}

}

func TestMarshalTransitivity(t *testing.T) {
	p := getTestingStruct()

	b, err := p.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}

	p2 := PaginationToken{}
	err = p2.UnmarshalBinary(b)
	if err != nil {
		t.Fatal(err)
	}

	if p.LastPaginatedType != p2.LastPaginatedType ||
		!bytes.Equal(p.LastKey, p2.LastKey) ||
		!p.TotalValue.Eq(p2.TotalValue) {
		t.Fatal("Should marshal to the same struct", p, p2)
	}

	b2, err := p2.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(b, b2) {
		t.Fatal("Should unmarshal to the same bytes", b, b2)
	}
}
