package crypto

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/crypto/bn256/cloudflare"
)

func TestSetPrivK(t *testing.T) {
	s := new(BNSigner)
	_, err := s.Pubkey()
	if err != ErrPrivkNotSet {
		t.Fatal("No error or incorrect error raised!")
	}
	privkBig, _ := new(big.Int).SetString("1234567890", 10)
	privkBig.Mod(privkBig, cloudflare.Order)
	privk := privkBig.Bytes()
	err = s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSignAndValidate(t *testing.T) {
	msg := []byte("some message")

	// Valid signature and validation
	s := &BNSigner{}
	_, err := s.Sign(msg)
	if err != ErrPrivkNotSet {
		t.Fatal("Should raise private key not set error!")
	}
	err = s.SetPrivk(Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	signature, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	v := &BNValidator{}
	valPubk, err := v.Validate(msg, signature)
	if err != nil {
		t.Fatal(err)
	}
	pubk, err := v.PubkeyFromSig(signature)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pubk, valPubk) {
		t.Fatal("pubkeys do not match!")
	}

	// Mess up signature
	signatureBad0 := signature[:96]
	_, err = v.Validate(msg, signatureBad0)
	if err == nil {
		t.Fatal("Improper signature should have raised an error")
	}

	// Have incorrect message for validation
	msgBad := []byte("another message")
	_, err = v.Validate(msgBad, signature)
	if err == nil {
		t.Fatal("Improper signature should have raised an error for incorrect public key")
	}
}

func TestSignerPubkeyFromSig(t *testing.T) {
	msg := []byte("some message")

	// Valid signature and validation
	s := &BNSigner{}
	err := s.SetPrivk(Hasher([]byte("secret")))
	if err != nil {
		t.Fatal(err)
	}
	signature, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	v := &BNValidator{}
	sigPubk, err := v.PubkeyFromSig(signature)
	if err != nil {
		t.Fatal(err)
	}
	signerPubk, err := s.Pubkey()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(sigPubk, signerPubk) {
		t.Fatalf("pubkeys do not match\n%x\n%x", sigPubk, signerPubk)
	}

	signatureBad0 := signature[:96]
	_, err = v.PubkeyFromSig(signatureBad0)
	if err == nil {
		t.Fatal("Should have raised error for invalid signature")
	}
}
