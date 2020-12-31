package cloudflare

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	privK1 := big.NewInt(1414213562373095) // 16 digits of sqrt(2)
	pubK1 := new(G1).ScalarBaseMult(privK1)
	part1 := 1
	secret1to2 := big.NewInt(1414213562373095) // 16 digits of sqrt(2)

	privK2 := big.NewInt(3141592653589793) // 16 digits of Pi
	pubK2 := new(G1).ScalarBaseMult(privK2)
	part2 := 2
	secret2to1 := big.NewInt(3141592653589793) // 16 digits of Pi

	encrypted1to2 := Encrypt(secret1to2, privK1, pubK2, part2)
	decrypted1to2 := Decrypt(encrypted1to2, privK2, pubK1, part2)
	if secret1to2.Cmp(decrypted1to2) != 0 {
		t.Fatal("Failed to correctly encrypt and decrypt shared secret 1 to 2")
	}

	encrypted2to1 := Encrypt(secret2to1, privK2, pubK1, part1)
	decrypted2to1 := Decrypt(encrypted2to1, privK1, pubK2, part1)
	if secret2to1.Cmp(decrypted2to1) != 0 {
		t.Fatal("Failed to correctly encrypt and decrypt shared secret 2 to 1")
	}

	// Now we decrypt using DecryptSS and the shared secret
	sharedSecret := new(G1).ScalarMult(pubK2, privK1)

	decrypted1to2SS := DecryptSS(encrypted1to2, sharedSecret, part2)
	if secret1to2.Cmp(decrypted1to2SS) != 0 {
		t.Fatal("DecryptSS failed to correctly decrypt shared secret 1 to 2")
	}

	decrypted2to1SS := DecryptSS(encrypted2to1, sharedSecret, part1)
	if secret2to1.Cmp(decrypted2to1SS) != 0 {
		t.Fatal("DecryptSS failed to correctly decrypt shared secret 2 to 1")
	}
}

func TestDLEQG1(t *testing.T) {
	k1 := big.NewInt(1)
	k2 := big.NewInt(123456789)
	alpha := big.NewInt(1414213562373095) // 16 digits of sqrt(2)
	x1 := new(G1).ScalarBaseMult(k1)
	x2 := new(G1).ScalarBaseMult(k2)
	y1 := new(G1).ScalarMult(x1, alpha)
	y2 := new(G1).ScalarMult(x2, alpha)

	pi, err0 := GenerateDLEQProofG1(x1, y1, x2, y2, alpha, rand.Reader)
	if err0 != nil {
		t.Fatal("Error occurred in GenDLEQProofG1")
	}

	err1 := VerifyDLEQProofG1(x1, y1, x2, y2, pi)
	if err1 != nil {
		t.Fatal("GenDLEQProofG1 failed to generate valid proof")
	}

	k3 := big.NewInt(2)
	y1Bad := new(G1).ScalarMult(x1, k3)
	err3 := VerifyDLEQProofG1(x1, y1Bad, x2, y2, pi)
	if err3 == nil {
		t.Fatal("VerifyDLEQG1 failed to raise error for invalid proof")
	}
}

func TestMarshalUint64ToUint256(t *testing.T) {
	// Test 0
	bytes0True := make([]byte, numBytes)
	bytes0 := marshalUint64ForUint256(0)
	if !bytes.Equal(bytes0True, bytes0) {
		t.Fatal("Should have equality for 0")
	}

	// Test 1
	bytes1True := make([]byte, numBytes)
	bytes1True[31] = 1
	bytes1 := marshalUint64ForUint256(1)
	if !bytes.Equal(bytes1True, bytes1) {
		t.Fatal("Should have equality for 1")
	}

	// Test 257 == 2^8 + 1
	bytes257True := make([]byte, numBytes)
	bytes257True[31] = 1
	bytes257True[30] = 1
	bytes257 := marshalUint64ForUint256(257)
	if !bytes.Equal(bytes257True, bytes257) {
		t.Fatal("Should have equality for 257")
	}

	// Test 65537 == 2^16 + 1
	bytes65537True := make([]byte, numBytes)
	bytes65537True[31] = 1
	bytes65537True[29] = 1
	bytes65537 := marshalUint64ForUint256(65537)
	if !bytes.Equal(bytes65537True, bytes65537) {
		t.Fatal("Should have equality for 65537")
	}
}
