package cloudflare

import (
	"testing"
)

func TestPairings(t *testing.T) {
	a1 := new(G1).ScalarBaseMult(bigFromBase10("1"))
	a2 := new(G1).ScalarBaseMult(bigFromBase10("2"))
	a37 := new(G1).ScalarBaseMult(bigFromBase10("37"))
	an1 := new(G1).ScalarBaseMult(bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495616"))

	b0 := new(G2).ScalarBaseMult(bigFromBase10("0"))
	b1 := new(G2).ScalarBaseMult(bigFromBase10("1"))
	b2 := new(G2).ScalarBaseMult(bigFromBase10("2"))
	b27 := new(G2).ScalarBaseMult(bigFromBase10("27"))
	b999 := new(G2).ScalarBaseMult(bigFromBase10("999"))
	bn1 := new(G2).ScalarBaseMult(bigFromBase10("21888242871839275222246405745257275088548364400416034343698204186575808495616"))

	p1 := Pair(a1, b1)
	pn1 := Pair(a1, bn1)
	np1 := Pair(an1, b1)
	if !pn1.IsEqual(np1) {
		t.Error("Pairing mismatch: e(a, -b) != e(-a, b)")
	}
	if !PairingCheck([]*G1{a1, an1}, []*G2{b1, b1}) {
		t.Error("MultiAte check gave false negative!")
	}
	p0 := new(GT).Add(p1, pn1)
	p0two := Pair(a1, b0)
	if !p0.IsEqual(p0two) {
		t.Error("Pairing mismatch: e(a, b) * e(a, -b) != 1")
	}
	p0three := new(GT).ScalarMult(p1, Order)
	if !p0.IsEqual(p0three) {
		t.Error("Pairing mismatch: e(a, b) has wrong order")
	}
	p2 := Pair(a2, b1)
	p2two := Pair(a1, b2)
	p2three := new(GT).ScalarMult(p1, bigFromBase10("2"))
	if !p2.IsEqual(p2two) {
		t.Error("Pairing mismatch: e(a, b * 2) != e(a * 2, b)")
	}
	if !p2.IsEqual(p2three) {
		t.Error("Pairing mismatch: e(a, b * 2) != e(a, b) ** 2")
	}
	if p2.IsEqual(p1) {
		t.Error("Pairing is degenerate!")
	}
	if PairingCheck([]*G1{a1, a1}, []*G2{b1, b1}) {
		t.Error("MultiAte check gave false positive!")
	}
	p999 := Pair(a37, b27)
	p999two := Pair(a1, b999)
	if !p999.IsEqual(p999two) {
		t.Error("Pairing mismatch: e(a * 37, b * 27) != e(a, b * 999)")
	}
}
