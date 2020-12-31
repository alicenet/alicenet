package secp256k1

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/secp256k1"
)

func makePrivateKey() *PrivateKey {
	privKey := make([]byte, 32)
	privKey[0] = 1
	privKey[31] = 1
	priv, _ := PrivKeyFromBytes(S256(), privKey)
	return priv
}

func makePrivateKey2() *PrivateKey {
	privKey := make([]byte, 32)
	privKey[0] = 1
	privKey[31] = 8
	priv, _ := PrivKeyFromBytes(S256(), privKey)
	return priv
}

func TestS256(t *testing.T) {
	curveParams := secp256k1.S256()
	cP := S256()

	if curveParams.N.Cmp(cP.N) != 0 {
		t.Fatal("Invalid N")
	}
	if curveParams.P.Cmp(cP.P) != 0 {
		t.Fatal("Invalid P")
	}
	if curveParams.B.Cmp(cP.B) != 0 {
		t.Fatal("Invalid B")
	}
	if curveParams.Gx.Cmp(cP.Gx) != 0 {
		t.Fatal("Invalid Gx")
	}
	if curveParams.Gy.Cmp(cP.Gy) != 0 {
		t.Fatal("Invalid Gy")
	}
	if curveParams.BitSize != cP.BitSize {
		t.Fatal("Invalid BitSize")
	}
}

func TestSerialize(t *testing.T) {
	curve := S256()
	privKeyTrue := make([]byte, 32)
	privKeyTrue[0] = 1
	privKeyTrue[31] = 1
	priv, _ := PrivKeyFromBytes(curve, privKeyTrue)
	privKey := priv.Serialize()
	if !bytes.Equal(privKeyTrue, privKey) {
		t.Fatal("privKeys do not agree")
	}
}

func TestNewPrivateKey(t *testing.T) {
	_, err := NewPrivateKey(S256())
	if err != nil {
		t.Fatal(err)
	}
}

func TestSerializeCompressed(t *testing.T) {
	priv := makePrivateKey()
	pubkeyTrue, err := hex.DecodeString("02e4dbb4350d84eabec1d67e40a398a78a8e6d719d86914393fca83b88dbe927af")
	if err != nil {
		t.Fatal(err)
	}
	pubkey := priv.PubKey().SerializeCompressed()
	if !bytes.Equal(pubkeyTrue, pubkey) {
		t.Fatal("pubkeys do not match (1)")
	}
	priv2 := makePrivateKey2()
	pubkey2True, err := hex.DecodeString("034bdfca3f5aea974acb970f52bbd570d88fd86fc3618ba367f4086bb87f3061b9")
	if err != nil {
		t.Fatal(err)
	}
	pubkey2 := priv2.PubKey().SerializeCompressed()
	if !bytes.Equal(pubkey2True, pubkey2) {
		t.Fatal("pubkeys do not match (2)")
	}
	nilPubkey := &PublicKey{}
	pubkey3 := nilPubkey.SerializeCompressed()
	zeroBytes := make([]byte, 0)
	if !bytes.Equal(pubkey3, zeroBytes) {
		fmt.Println(pubkey3)
		t.Fatal("Should have returned slice of zero bytes")
	}
}

func TestParsePubKey(t *testing.T) {
	badBytes := make([]byte, 0)
	curve := S256()
	_, err := ParsePubKey(badBytes, curve)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	badBytes2 := make([]byte, 1)
	_, err = ParsePubKey(badBytes2, curve)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	badBytes3 := make([]byte, pubkeyBytesLenCompressed)
	_, err = ParsePubKey(badBytes3, curve)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	badBytes4 := make([]byte, pubkeyBytesLenCompressed)
	badBytes4[0] = 0x02
	_, err = ParsePubKey(badBytes4, curve)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	priv := makePrivateKey()
	pubTrue := priv.PubKey()
	pubkeyBytes := pubTrue.SerializeCompressed()
	pub, err := ParsePubKey(pubkeyBytes, curve)
	if err != nil {
		t.Fatal(err)
	}
	if pubTrue.X.Cmp(pub.X) != 0 || pubTrue.Y.Cmp(pub.Y) != 0 {
		t.Fatal("pubkeys do not match (1)")
	}

	priv2 := makePrivateKey2()
	pub2True := priv2.PubKey()
	pubkey2Bytes := pub2True.SerializeCompressed()
	pub2, err := ParsePubKey(pubkey2Bytes, curve)
	if err != nil {
		t.Fatal(err)
	}
	if pub2True.X.Cmp(pub2.X) != 0 || pub2True.Y.Cmp(pub2.Y) != 0 {
		t.Fatal("pubkeys do not match (2)")
	}
}

func TestPrivKeyFromBytes(t *testing.T) {
	privKey := make([]byte, 32)
	privKey[0] = 1
	privKey[31] = 1
	priv, _ := PrivKeyFromBytes(S256(), privKey)
	privTrue := makePrivateKey()
	if privTrue.D.Cmp(priv.D) != 0 {
		t.Fatal("private key does not match")
	}
}

func TestDecompressPoint(t *testing.T) {
	x := new(big.Int)
	curve := S256()
	yBit := true
	_, err := decompressPoint(curve, x, yBit)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}
	yBit = false
	_, err = decompressPoint(curve, x, yBit)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	_, err = decompressPoint(nil, x, yBit)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	_, err = decompressPoint(curve, nil, yBit)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}

	yBit = true
	x = big.NewInt(1)
	_, err = decompressPoint(curve, x, yBit)
	if err != nil {
		t.Fatal(err)
	}
	yBit = false
	_, err = decompressPoint(curve, x, yBit)
	if err != nil {
		t.Fatal(err)
	}

	yBit = true
	x = big.NewInt(5)
	_, err = decompressPoint(curve, x, yBit)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}
}

func TestIsOnCurve(t *testing.T) {
	curve := S256()
	if !isOnCurve(curve, curve.Gx, curve.Gy) {
		t.Fatal("Should have base point on curve")
	}
	big1 := big.NewInt(1)
	big2 := big.NewInt(2)
	if isOnCurve(curve, big1, big2) {
		t.Fatal("Should not have point on curve")
	}
	if isOnCurve(nil, big1, big2) {
		t.Fatal("Should have raised error (1)")
	}
	if isOnCurve(curve, nil, big2) {
		t.Fatal("Should have raised error (2)")
	}
	if isOnCurve(curve, big1, nil) {
		t.Fatal("Should have raised error (3)")
	}
	priv := makePrivateKey()
	if !isOnCurve(curve, priv.PubKey().X, priv.PubKey().Y) {
		t.Fatal("Should have public key on curve (1)")
	}
	priv = makePrivateKey2()
	if !isOnCurve(curve, priv.PubKey().X, priv.PubKey().Y) {
		t.Fatal("Should have public key on curve (2)")
	}
}

func TestIsOdd(t *testing.T) {
	big1 := big.NewInt(1)
	if !isOdd(big1) {
		t.Fatal("Should return 1 as odd")
	}
	big65537 := big.NewInt(65537)
	if !isOdd(big65537) {
		t.Fatal("Should return 65537 as odd")
	}
	big2 := big.NewInt(2)
	if isOdd(big2) {
		t.Fatal("Should return 2 as not odd")
	}
	big0 := new(big.Int)
	if isOdd(big0) {
		t.Fatal("Should return 0 as not odd")
	}
}
