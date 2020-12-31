package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestSecpPrivkPubkey(t *testing.T) {
	s := new(Secp256k1Signer)
	_, err := s.Pubkey()
	if err != ErrPrivkNotSet {
		t.Fatal("Should have raised error for secp256k1 privk not set!")
	}
	privkBad := make([]byte, 31)
	privkBad[0] = 1
	privkBad[30] = 1
	err = s.SetPrivk(privkBad)
	if err == nil {
		t.Fatal("Error should have been raised for not enough bits")
	}
	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	err = s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	pubk, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	pubkTrue, err := hex.DecodeString("04e4dbb4350d84eabec1d67e40a398a78a8e6d719d86914393fca83b88dbe927afb80fe66bf659859889a544623c945d0bd80d855f649e8c197be3aa41fe0390f8")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pubk, pubkTrue) {
		t.Fatal("pubk does not match true value!")
	}
}

func TestSecpSign(t *testing.T) {
	s := new(Secp256k1Signer)
	msg := []byte("A message to sign")
	_, err := s.Sign(msg)
	if err == nil {
		t.Fatal("Error should be raised due to no privk!")
	}
	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	err = s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	digestHash := Hasher(msg)
	_, err = s.Sign(digestHash)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSecpPubkeyFromSig(t *testing.T) {
	s := new(Secp256k1Signer)
	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	err := s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("A message to sign")
	sig, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	v := new(Secp256k1Validator)
	pubk, err := v.PubkeyFromSig(msg, sig)
	if err != nil {
		t.Fatal(err)
	}
	pubkTrue, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pubkTrue, pubk) {
		t.Fatalf("pubks do not agree!\n%x\n%x\n", pubkTrue, pubk)
	}
	// Make signature too short
	sigBad0 := []byte{}
	_, err = v.PubkeyFromSig(msg, sigBad0)
	if err != ErrInvalidSignature {
		t.Fatal("Should have raised error for signature being too short!")
	}
	// Make public key invalid
	sigBad1 := sig
	sigBad1[0] = 1
	badpubk, err := v.PubkeyFromSig(msg, sigBad1)
	if err == nil {
		t.Fatal("Should not have been able to recover publickey.")
	}
	if bytes.Equal(badpubk, pubkTrue) {
		t.Fatal("Should not have correct public key for invalid signature!")
	}
}

func TestSecpValidate(t *testing.T) {
	s := new(Secp256k1Signer)
	privk := make([]byte, 32)
	privk[0] = 1
	privk[31] = 1
	err := s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	pubkTrue, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	msg := []byte("A message to sign")
	sig, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	v := new(Secp256k1Validator)
	pubk, err := v.Validate(msg, sig)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pubk, pubkTrue) {
		t.Fatal("pubks do not match!")
	}
	// Make full signature too short
	sigBad0 := sig[:32]
	_, err = v.Validate(msg, sigBad0)
	if err != ErrInvalidSignature {
		t.Fatal("Should have raised error for signature being incorrect length!")
	}
	// Change sig portion of full signature
	sigBad1 := sig[:64]
	_, err = v.Validate(msg, sigBad1)
	if err == nil {
		t.Fatal("Error should be raised for invalid signature!")
	}
	// Make public key portion of signature not match public key from ECRecover
	sigBad2 := sig
	sigBad2[0] = 9
	badpubk, err := v.Validate(msg, sigBad2)
	if err == nil {
		t.Fatal("Should not have been able to recover publickey.")
	}
	if bytes.Equal(badpubk, pubkTrue) {
		t.Fatal("Should not have correct public key for invalid signature!")
	}
}
