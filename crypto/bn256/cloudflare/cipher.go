package cloudflare

import (
	"crypto/rand"
	"io"
	"math/big"
)

// Encrypt uses the individual's private key privK and participantIndex's
// public key pubK to encrypt secretValue to encryptedValue.
// Encryption and decryption are based on using the hash of the shared secret
// from Diffie-Hellman Key Exchange along with the index of the recipient
// receiving the message as a one-time pad.
// Including the receiver's index ensures it is only used once.
func Encrypt(secretValue, privK *big.Int, pubK *G1, participantIndex int) *big.Int {
	sharedSecret := new(G1).ScalarMult(pubK, privK)
	sharedSecretSlice := sharedSecret.Marshal()
	sharedSecretSlice = sharedSecretSlice[:numBytes] // Only use x-coordinate
	indexBytes := marshalUint64ForUint256(uint64(participantIndex))
	byteHash := append(sharedSecretSlice, indexBytes...)
	hashOutput := HashFunc256(byteHash)
	oneTimePad := new(big.Int).SetBytes(hashOutput[:])
	encryptedValue := new(big.Int).Xor(oneTimePad, secretValue)
	return encryptedValue
}

// Decrypt uses privK (participantIndex is the index who called this function)
// and pubK (who sent encryptedValue) to decrypt encryptedValue to secretValue
// using the shared secret.
// Encryption and decryption are based on using the hash of the shared secret
// from Diffie-Hellman Key Exchange along with the index of the recipient
// receiving the message as a one-time pad.
// Including the receiver's index ensures it is only used once.
func Decrypt(encryptedValue, privK *big.Int, pubK *G1, participantIndex int) *big.Int {
	sharedSecret := new(G1).ScalarMult(pubK, privK)
	sharedSecretSlice := sharedSecret.Marshal()
	sharedSecretSlice = sharedSecretSlice[:numBytes] // Only use x-coordinate
	indexBytes := marshalUint64ForUint256(uint64(participantIndex))
	byteHash := append(sharedSecretSlice, indexBytes...)
	hashOutput := HashFunc256(byteHash)
	oneTimePad := new(big.Int).SetBytes(hashOutput[:])
	secretValue := new(big.Int).Xor(oneTimePad, encryptedValue)
	return secretValue
}

// marshalUint64ForUint256 converts uint64 into big.Int and stores it in
// big-endian format for 256-bit unsigned integer.
// We will not run into issues because xBytesLen <= 8
// (due to fact j is uint64) and byteSlice is 32 bytes.
//
// This is required because this will be evaluated by the Ethereum Virtual
// Machine, which uses uint256 as its standard unsigned integer.
func marshalUint64ForUint256(j uint64) []byte {
	x := new(big.Int).SetUint64(j)
	xBytes := x.Bytes()
	xBytesLen := len(xBytes)
	byteSlice := make([]byte, numBytes)
	for j := 1; j <= xBytesLen; j++ {
		byteSlice[numBytes-j] = xBytes[xBytesLen-j]
	}
	return byteSlice
}

// DecryptSS uses the x coordinate of the shared secret kX to
// decrypt encryptedValue to secretValue. See Decrypt for more information.
func DecryptSS(encryptedValue *big.Int, sharedSecret *G1, participantIndex int) *big.Int {
	sharedSecretSlice := sharedSecret.Marshal()
	sharedSecretSlice = sharedSecretSlice[:numBytes] // Only use x-coordinate
	indexBytes := marshalUint64ForUint256(uint64(participantIndex))
	byteHash := append(sharedSecretSlice, indexBytes...)
	hashOutput := HashFunc256(byteHash)
	oneTimePad := new(big.Int).SetBytes(hashOutput[:])
	secretValue := new(big.Int).Xor(oneTimePad, encryptedValue)
	return secretValue
}

// GenerateDLEQProofG1 generates the discrete log equality proof, showing
//
// 		y1 == x1^alpha and y2 == x2^alpha
//
// without disclosing alpha. It is based on the premise that if w is random and
//
//		t1 == x1^w    and    t2 == x2^w,
//
// then
//
//		x1^r * y1^c == x1^w == t1    and    x2^r * y2^c == x2^w == t2,
//
// where c and r are chosen beforehand to satisfy
//
//		c == Hash(x1, y1, x2, y2, t1, t2)
//
// and
//
//		r == w - alpha*c.
//
// This is used during the distributed key generation protocol to ensure
// honest participation.
func GenerateDLEQProofG1(x1, y1, x2, y2 *G1, alpha *big.Int, rIO io.Reader) ([2]*big.Int, error) {
	w, err := rand.Int(rIO, Order)
	if (err != nil) || (w.Cmp(big.NewInt(0)) == 0) {
		panic("GenerateDLEQProofG1: Error in generating random number for zk-proof")
	}
	t1 := new(G1).ScalarMult(x1, w)
	t2 := new(G1).ScalarMult(x2, w)

	x1Bytes := x1.Marshal()
	y1Bytes := y1.Marshal()
	x2Bytes := x2.Marshal()
	y2Bytes := y2.Marshal()
	t1Bytes := t1.Marshal()
	t2Bytes := t2.Marshal()

	byteHash := make([]byte, 0, 384) // 384 == 12*32
	byteHash = append(byteHash, x1Bytes...)
	byteHash = append(byteHash, y1Bytes...)
	byteHash = append(byteHash, x2Bytes...)
	byteHash = append(byteHash, y2Bytes...)
	byteHash = append(byteHash, t1Bytes...)
	byteHash = append(byteHash, t2Bytes...)

	cBytes := HashFunc256(byteHash)
	c := new(big.Int).SetBytes(cBytes[:])
	r := new(big.Int)
	r.Mul(alpha, c)
	r.Neg(r)
	r.Add(r, w)
	r.Mod(r, Order)
	pi := [2]*big.Int{c, r}
	return pi, nil
}

// VerifyDLEQProofG1 verifies the discrete log proof from GenerateDLEQProofG1;
// returns nil for a valid proof and an error when invalid.
// This verifies that
//
// 		y1 == x1^alpha    and    y2 == x2^alpha
//
// without disclosing alpha. From the proof pi == (c, r),
// if the above equalities hold, then
//
//		x1^r * y1^c == x1^w == t1    and    x2^r * y2^c == x2^w == t2,
//
// where w is a random value found in generating the proof.
// This is used during the distributed key generation protocol to ensure
// honest participation.
func VerifyDLEQProofG1(x1, y1, x2, y2 *G1, pi [2]*big.Int) error {
	c := new(big.Int).Set(pi[0])
	r := new(big.Int).Set(pi[1])

	x1Pow := new(G1).ScalarMult(x1, r)
	y1Pow := new(G1).ScalarMult(y1, c)
	x2Pow := new(G1).ScalarMult(x2, r)
	y2Pow := new(G1).ScalarMult(y2, c)
	t1Prime := new(G1).Add(x1Pow, y1Pow)
	t2Prime := new(G1).Add(x2Pow, y2Pow)

	x1Bytes := x1.Marshal()
	y1Bytes := y1.Marshal()
	x2Bytes := x2.Marshal()
	y2Bytes := y2.Marshal()
	t1PrimeBytes := t1Prime.Marshal()
	t2PrimeBytes := t2Prime.Marshal()

	byteHash := make([]byte, 0, 384) // 384 == 12*32
	byteHash = append(byteHash, x1Bytes...)
	byteHash = append(byteHash, y1Bytes...)
	byteHash = append(byteHash, x2Bytes...)
	byteHash = append(byteHash, y2Bytes...)
	byteHash = append(byteHash, t1PrimeBytes...)
	byteHash = append(byteHash, t2PrimeBytes...)

	cPrimeBytes := HashFunc256(byteHash)
	cPrime := new(big.Int).SetBytes(cPrimeBytes[:])
	if cPrime.Cmp(c) != 0 {
		return ErrDLEQInvalidProof
	}
	return nil
}
