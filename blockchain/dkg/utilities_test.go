package dkg_test

import (
	"math/big"
	"testing"

	"github.com/MadBase/MadNet/blockchain/dkg"
	"github.com/holiman/uint256"
)

func TestFoo(t *testing.T) {
	a := uint256.NewInt(4)
	b := uint256.NewInt(6).SetAllOne()
	c := uint256.NewInt(8)
	var d *uint256.Int

	var overflow bool

	t.Logf("a:%v b:%v c:%v d:%v overflow:%v", a.String(), b.String(), c.String(), nil, overflow)

	d, overflow = c.AddOverflow(a, b)
	t.Logf("a:%v b:%v c:%v d:%v overflow:%v", a.String(), b.String(), c.String(), d.String(), overflow)
	t.Logf("ptr c:%p d:%p", c, d)
}

func TestIntsToBigInts(t *testing.T) {
	ints := []int{1, 2, 3, 5, 8, 13}
	big1 := big.NewInt(1)
	big2 := big.NewInt(2)
	big3 := big.NewInt(3)
	big5 := big.NewInt(5)
	big8 := big.NewInt(8)
	big13 := big.NewInt(13)
	bigIntsTrue := []*big.Int{}
	bigIntsTrue = append(bigIntsTrue, big1)
	bigIntsTrue = append(bigIntsTrue, big2)
	bigIntsTrue = append(bigIntsTrue, big3)
	bigIntsTrue = append(bigIntsTrue, big5)
	bigIntsTrue = append(bigIntsTrue, big8)
	bigIntsTrue = append(bigIntsTrue, big13)

	bigInts := dkg.IntsToBigInts(ints)
	if len(bigInts) != len(bigIntsTrue) {
		t.Fatal("Invalid return length")
	}

	for i := 0; i < len(ints); i++ {
		bigIntValue := bigInts[i]
		bigIntTrueValue := bigIntsTrue[i]
		if bigIntValue.Cmp(bigIntTrueValue) != 0 {
			t.Fatal("Invalid bigInt value")
		}
	}
}
