package crypto

import (
	"bytes"
	"math/big"
	"testing"

	bn256 "github.com/alicenet/alicenet/crypto/bn256/cloudflare"
)

func TestGroupSignerSetPrivK(t *testing.T) {
	s := new(BNGroupSigner)
	privkBig, _ := new(big.Int).SetString("1234567890", 10)
	privkBig.Mod(privkBig, bn256.Order)
	privk := privkBig.Bytes()
	err := s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGroupSignerSetGroupPubk(t *testing.T) {
	s := new(BNGroupSigner)
	msk := big.NewInt(1234567890)
	groupPubkBN := new(bn256.G2).ScalarBaseMult(msk)
	groupPubk := groupPubkBN.Marshal()
	err := s.SetGroupPubk(groupPubk)
	if err != nil {
		t.Fatal(err)
	}

	// This will fail
	groupPubkBad := groupPubk[:96]
	err = s.SetGroupPubk(groupPubkBad)
	if err == nil {
		t.Fatal("Error should have been raised for invalid group master public key!")
	}
}

func TestGroupSignerPubkeyShare(t *testing.T) {
	s := new(BNGroupSigner)
	// Portion will fail because not initialized
	_, err := s.PubkeyShare()
	if err != ErrPrivkNotSet {
		t.Fatal("No public key should exist!")
	}

	// This should work because we first initialize
	privkBig, _ := new(big.Int).SetString("1234567890", 10)
	privkBig.Mod(privkBig, bn256.Order)
	privk := privkBig.Bytes()
	pubkBN := new(bn256.G2).ScalarBaseMult(privkBig)
	pubkBNBytes := pubkBN.Marshal()
	err = s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	pubkBytes, err := s.PubkeyShare()
	if err != nil {
		t.Fatal("Error occurred when calling bns.PubkeyShare()")
	}
	if !bytes.Equal(pubkBytes, pubkBNBytes) {
		t.Fatal("pubkey bytes are not equal!")
	}
}

func TestGroupSignerPubkeyGroup(t *testing.T) {
	s := new(BNGroupSigner)
	// Portion will fail because not initialized
	_, err := s.PubkeyGroup()
	if err == nil {
		t.Fatal("No group public key should exist!")
	}

	msk := big.NewInt(1234567890)
	groupPubkBN := new(bn256.G2).ScalarBaseMult(msk)
	groupPubk := groupPubkBN.Marshal()
	err = s.SetGroupPubk(groupPubk)
	if err != nil {
		t.Fatal(err)
	}
	groupPubkReturned, err := s.PubkeyGroup()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(groupPubkReturned, groupPubk) {
		t.Fatal("groupPubk's do not match!")
	}
}

func TestGroupSignAndValidate(t *testing.T) {
	msg := []byte("A message to sign")

	// Create signature
	s := new(BNGroupSigner)
	privkBig := big.NewInt(1234567890)
	privk := privkBig.Bytes()
	err := s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	pubkTrue, err := s.PubkeyShare()
	if err != nil {
		t.Fatal("Error when calling s.PubkeyShare")
	}

	// Test valid signature from above
	v := new(BNGroupValidator)
	pubk, err := v.Validate(msg, sig)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pubk, pubkTrue) {
		t.Fatal("pubks do not match!")
	}

	// Mess up signature
	sigBad0 := sig[:96]
	_, err = v.Validate(msg, sigBad0)
	if err == nil {
		t.Fatal("Error should have been raised for invalid signature")
	}

	// Change message, invalidating signature
	msgBad := []byte("Another message to sign")
	_, err = v.Validate(msgBad, sig)
	if err == nil {
		t.Fatal("Error should have been raised; incorrect public key in sig")
	}
}

func TestGroupSignerPubkeyFromSig(t *testing.T) {
	msg := []byte("A message to sign")

	// Create signature
	s := new(BNGroupSigner)
	privkBig := big.NewInt(1234567890)
	privk := privkBig.Bytes()
	err := s.SetPrivk(privk)
	if err != nil {
		t.Fatal(err)
	}
	sig, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	pubkeyTrue, err := s.PubkeyShare()
	if err != nil {
		t.Fatal(err)
	}

	// Correctly generate public key
	v := new(BNGroupValidator)
	pubkey, err := v.PubkeyFromSig(sig)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(pubkey, pubkeyTrue) {
		t.Fatal("pubkey from sig does not match pubkey from signer!")
	}

	// Change sig to result in invalid public key
	sigBad := sig[:96]
	_, err = v.PubkeyFromSig(sigBad)
	if err == nil {
		t.Fatal("Error should have been raised for invalid signature!")
	}
}

func TestVerifyGroupShares(t *testing.T) {
	secret1 := big.NewInt(100)
	secret2 := big.NewInt(101)
	secret3 := big.NewInt(102)
	secret4 := big.NewInt(103)

	msk := big.NewInt(0)
	msk.Add(msk, secret1)
	msk.Add(msk, secret2)
	msk.Add(msk, secret3)
	msk.Add(msk, secret4)
	msk.Mod(msk, bn256.Order)
	mpk := new(bn256.G2).ScalarBaseMult(msk)

	big1 := big.NewInt(1)
	big2 := big.NewInt(2)

	privCoefs1 := []*big.Int{secret1, big1, big2}
	privCoefs2 := []*big.Int{secret2, big1, big2}
	privCoefs3 := []*big.Int{secret3, big1, big2}
	privCoefs4 := []*big.Int{secret4, big1, big2}

	share1to1 := bn256.PrivatePolyEval(privCoefs1, 1)
	share1to2 := bn256.PrivatePolyEval(privCoefs1, 2)
	share1to3 := bn256.PrivatePolyEval(privCoefs1, 3)
	share1to4 := bn256.PrivatePolyEval(privCoefs1, 4)
	share2to1 := bn256.PrivatePolyEval(privCoefs2, 1)
	share2to2 := bn256.PrivatePolyEval(privCoefs2, 2)
	share2to3 := bn256.PrivatePolyEval(privCoefs2, 3)
	share2to4 := bn256.PrivatePolyEval(privCoefs2, 4)
	share3to1 := bn256.PrivatePolyEval(privCoefs3, 1)
	share3to2 := bn256.PrivatePolyEval(privCoefs3, 2)
	share3to3 := bn256.PrivatePolyEval(privCoefs3, 3)
	share3to4 := bn256.PrivatePolyEval(privCoefs3, 4)
	share4to1 := bn256.PrivatePolyEval(privCoefs4, 1)
	share4to2 := bn256.PrivatePolyEval(privCoefs4, 2)
	share4to3 := bn256.PrivatePolyEval(privCoefs4, 3)
	share4to4 := bn256.PrivatePolyEval(privCoefs4, 4)

	groupShares := make([][]byte, 4)
	for k := 0; k < len(groupShares); k++ {
		groupShares[k] = make([]byte, len(mpk.Marshal()))
	}

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := bn256.GenerateGroupSecretKeyPortion(listOfSS1)
	gpk1 := new(bn256.G2).ScalarBaseMult(gsk1)
	groupShares[0] = gpk1.Marshal()

	listOfSS2 := []*big.Int{share1to2, share2to2, share3to2, share4to2}
	gsk2 := bn256.GenerateGroupSecretKeyPortion(listOfSS2)
	gpk2 := new(bn256.G2).ScalarBaseMult(gsk2)
	groupShares[1] = gpk2.Marshal()

	listOfSS3 := []*big.Int{share1to3, share2to3, share3to3, share4to3}
	gsk3 := bn256.GenerateGroupSecretKeyPortion(listOfSS3)
	gpk3 := new(bn256.G2).ScalarBaseMult(gsk3)
	groupShares[2] = gpk3.Marshal()

	listOfSS4 := []*big.Int{share1to4, share2to4, share3to4, share4to4}
	gsk4 := bn256.GenerateGroupSecretKeyPortion(listOfSS4)
	gpk4 := new(bn256.G2).ScalarBaseMult(gsk4)
	groupShares[3] = gpk4.Marshal()

	// Test good version
	s := new(BNGroupSigner)
	err := s.VerifyGroupShares(groupShares)
	if err != nil {
		t.Fatal("Something went wrong with bns.SetGroupShares; should work")
	}

	// Test bad versions
	groupSharesBad0 := make([][]byte, 5)
	for k := 0; k < len(groupSharesBad0); k++ {
		groupSharesBad0[k] = make([]byte, len(mpk.Marshal()))
	}
	groupSharesBad0[0] = gpk1.Marshal()
	groupSharesBad0[1] = gpk2.Marshal()
	groupSharesBad0[2] = gpk3.Marshal()
	groupSharesBad0[3] = gpk4.Marshal()
	groupSharesBad0[4] = gpk1.Marshal()
	err = s.VerifyGroupShares(groupSharesBad0)
	if err != ErrInvalidPubkeyShares {
		t.Fatal("Should raise error for having repeated public keys")
	}

	groupSharesBad1 := make([][]byte, 4)
	for k := 0; k < len(groupSharesBad1); k++ {
		groupSharesBad1[k] = make([]byte, len(mpk.Marshal()))
	}
	groupSharesBad1[0] = gpk1.Marshal()
	groupSharesBad1[1] = gpk2.Marshal()
	groupSharesBad1[2] = gpk3.Marshal()
	groupSharesBad1[3] = gpk4.Marshal()
	for k := 0; k < 32; k++ {
		groupSharesBad1[0][k] = 0
	}
	err = s.VerifyGroupShares(groupSharesBad1)
	if err == nil {
		t.Fatal("Should raise error for having invalid public key")
	}
}

func TestGroupSignerAggregate(t *testing.T) {
	s := new(BNGroupSigner)
	msg := []byte("A message to sign")

	secret1 := big.NewInt(100)
	secret2 := big.NewInt(101)
	secret3 := big.NewInt(102)
	secret4 := big.NewInt(103)

	msk := big.NewInt(0)
	msk.Add(msk, secret1)
	msk.Add(msk, secret2)
	msk.Add(msk, secret3)
	msk.Add(msk, secret4)
	msk.Mod(msk, bn256.Order)
	mpk := new(bn256.G2).ScalarBaseMult(msk)

	big1 := big.NewInt(1)
	big2 := big.NewInt(2)

	privCoefs1 := []*big.Int{secret1, big1, big2}
	privCoefs2 := []*big.Int{secret2, big1, big2}
	privCoefs3 := []*big.Int{secret3, big1, big2}
	privCoefs4 := []*big.Int{secret4, big1, big2}

	share1to1 := bn256.PrivatePolyEval(privCoefs1, 1)
	share1to2 := bn256.PrivatePolyEval(privCoefs1, 2)
	share1to3 := bn256.PrivatePolyEval(privCoefs1, 3)
	share1to4 := bn256.PrivatePolyEval(privCoefs1, 4)
	share2to1 := bn256.PrivatePolyEval(privCoefs2, 1)
	share2to2 := bn256.PrivatePolyEval(privCoefs2, 2)
	share2to3 := bn256.PrivatePolyEval(privCoefs2, 3)
	share2to4 := bn256.PrivatePolyEval(privCoefs2, 4)
	share3to1 := bn256.PrivatePolyEval(privCoefs3, 1)
	share3to2 := bn256.PrivatePolyEval(privCoefs3, 2)
	share3to3 := bn256.PrivatePolyEval(privCoefs3, 3)
	share3to4 := bn256.PrivatePolyEval(privCoefs3, 4)
	share4to1 := bn256.PrivatePolyEval(privCoefs4, 1)
	share4to2 := bn256.PrivatePolyEval(privCoefs4, 2)
	share4to3 := bn256.PrivatePolyEval(privCoefs4, 3)
	share4to4 := bn256.PrivatePolyEval(privCoefs4, 4)

	groupShares := make([][]byte, 4)
	for k := 0; k < len(groupShares); k++ {
		groupShares[k] = make([]byte, len(mpk.Marshal()))
	}

	listOfSS1 := []*big.Int{share1to1, share2to1, share3to1, share4to1}
	gsk1 := bn256.GenerateGroupSecretKeyPortion(listOfSS1)
	gpk1 := new(bn256.G2).ScalarBaseMult(gsk1)
	groupShares[0] = gpk1.Marshal()
	s1 := new(BNGroupSigner)
	err := s1.SetPrivk(gsk1.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig1, err := s1.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	listOfSS2 := []*big.Int{share1to2, share2to2, share3to2, share4to2}
	gsk2 := bn256.GenerateGroupSecretKeyPortion(listOfSS2)
	gpk2 := new(bn256.G2).ScalarBaseMult(gsk2)
	groupShares[1] = gpk2.Marshal()
	s2 := new(BNGroupSigner)
	err = s2.SetPrivk(gsk2.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig2, err := s2.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	listOfSS3 := []*big.Int{share1to3, share2to3, share3to3, share4to3}
	gsk3 := bn256.GenerateGroupSecretKeyPortion(listOfSS3)
	gpk3 := new(bn256.G2).ScalarBaseMult(gsk3)
	groupShares[2] = gpk3.Marshal()
	s3 := new(BNGroupSigner)
	err = s3.SetPrivk(gsk3.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig3, err := s3.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	listOfSS4 := []*big.Int{share1to4, share2to4, share3to4, share4to4}
	gsk4 := bn256.GenerateGroupSecretKeyPortion(listOfSS4)
	gpk4 := new(bn256.G2).ScalarBaseMult(gsk4)
	groupShares[3] = gpk4.Marshal()
	s4 := new(BNGroupSigner)
	err = s4.SetPrivk(gsk4.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	sig4, err := s4.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}

	sigs := make([][]byte, 4)
	for k := 0; k < len(sigs); k++ {
		sigs[k] = make([]byte, 192)
	}
	sigs[0] = sig1
	sigs[1] = sig2
	sigs[2] = sig3
	sigs[3] = sig4

	// Attempt with empty GroupShares
	emptyShares := make([][]byte, 4)
	_, err = s.Aggregate(sigs, emptyShares)
	if err == nil {
		t.Fatal("Error should have been raised for invalid groupShares!")
	}

	// Attempt without groupPubk
	_, err = s.Aggregate(sigs, groupShares)
	if err != ErrPubkeyGroupNotSet {
		t.Fatal("Error should have been raised for no PubkeyGroup!")
	}
	err = s.SetGroupPubk(mpk.Marshal())
	if err != nil {
		t.Fatal(err)
	}

	// Make bad sigs array
	sigsBad := make([][]byte, 2)
	for k := 0; k < len(sigsBad); k++ {
		sigsBad[k] = make([]byte, 192)
	}
	sigsBad[0] = sig1
	sigsBad[1] = sig2
	_, err = s.Aggregate(sigsBad, groupShares)
	if err == nil {
		t.Fatal("Should have raised an error for too few signatures!")
	}

	// Finally submit signature
	grpsig, err := s.Aggregate(sigs, groupShares)
	if err != nil {
		t.Fatal(err)
	}
	v := new(BNGroupValidator)
	grpsigReturned, err := v.Validate(msg, grpsig)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(mpk.Marshal(), grpsigReturned) {
		t.Fatal("grpsigs do not match!")
	}
}
