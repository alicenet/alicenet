package dkg_test

import (
	"testing"

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
