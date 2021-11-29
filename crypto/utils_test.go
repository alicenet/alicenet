package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestGetAccount(t *testing.T) {
	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	s := &Secp256k1Signer{}
	err := s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	pubk, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	acctTrue, err := hex.DecodeString("7691ee0343b9a529675e1a8a70197b3b704f90b7")
	if err != nil {
		t.Fatal(err)
	}
	acct := GetAccount(pubk)
	if !bytes.Equal(acct, acctTrue) {
		t.Fatal("acct does not match true value!")
	}
}

func TestCalcThreshold(t *testing.T) {
	n := 4
	threshold := CalcThreshold(n)
	if threshold != 2 {
		t.Fatal("We should have t == 2!")
	}
	n = 5
	threshold = CalcThreshold(n)
	if threshold != 3 {
		t.Fatal("We should have t == 3!")
	}
	n = 6
	threshold = CalcThreshold(n)
	if threshold != 4 {
		t.Fatal("We should have t == 4!")
	}
	n = 7
	threshold = CalcThreshold(n)
	if threshold != 4 {
		t.Fatal("We should have t == 4!")
	}
	n = 8
	threshold = CalcThreshold(n)
	if threshold != 5 {
		t.Fatal("We should have t == 5!")
	}
	n = 9
	threshold = CalcThreshold(n)
	if threshold != 6 {
		t.Fatal("We should have t == 6!")
	}
}
