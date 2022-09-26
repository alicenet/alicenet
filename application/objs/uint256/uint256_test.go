package uint256

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/utils"
)

type testObj struct {
	Value *Uint256
}

func TestUint256MarshalBinary(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	_, err = u.MarshalBinary()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	zeroBytes := make([]byte, 32)
	u1 := &Uint256{}
	u1.SetZero()
	ret, err := u1.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ret, zeroBytes) {
		t.Fatal("Did not return zero byte slice")
	}

	v := &Uint256{}
	oneBytes := make([]byte, 32)
	oneBytes[31] = 1
	v.SetOne()
	ret, err = v.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ret, oneBytes) {
		t.Fatal("Did not return one byte slice")
	}
}

func TestUint256UnmarshalBinary(t *testing.T) {
	zeroBytes := make([]byte, 32)
	obj := &testObj{}
	err := obj.Value.UnmarshalBinary(zeroBytes)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	err = u.UnmarshalBinary(zeroBytes)
	if err != nil {
		t.Fatal(err)
	}
	if !u.IsZero() {
		t.Fatal("Should equal 0")
	}

	data := make([]byte, 32)
	data[0] = 0
	data[1] = 1
	data[2] = 2
	data[3] = 3
	data[4] = 4
	data[5] = 5
	data[6] = 6
	data[7] = 7
	data[8] = 8
	data[9] = 9
	data[10] = 10
	data[11] = 11
	data[12] = 12
	data[13] = 13
	data[14] = 14
	data[15] = 15
	data[16] = 16
	data[17] = 17
	data[18] = 18
	data[19] = 19
	data[20] = 20
	data[21] = 21
	data[22] = 22
	data[23] = 23
	data[24] = 24
	data[25] = 25
	data[26] = 26
	data[27] = 27
	data[28] = 28
	data[29] = 29
	data[30] = 30
	data[31] = 31
	err = u.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := u.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ret, data) {
		t.Fatal("byte slices do not match")
	}

	nilBytes := make([]byte, 0)
	err = u.UnmarshalBinary(nilBytes)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	tooLongBytes := make([]byte, 33)
	err = u.UnmarshalBinary(tooLongBytes)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
}

func TestUint256MarshalString(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.MarshalString()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	_, err = u.MarshalString()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	u.SetZero()
	uString, err := u.MarshalString()
	if err != nil {
		t.Fatal(err)
	}
	uStringTrue := "0000000000000000000000000000000000000000000000000000000000000000"
	if uString != uStringTrue {
		t.Fatal("strings do not match")
	}
}

func TestUint256UnmarshalString(t *testing.T) {
	zeroString := "0000000000000000000000000000000000000000000000000000000000000000"
	obj := &testObj{}
	err := obj.Value.UnmarshalString(zeroString)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	err = u.UnmarshalString(zeroString)
	if err != nil {
		t.Fatal(err)
	}
	if !u.IsZero() {
		t.Fatal("Incorrect value (1)")
	}

	zeroString2 := "0"
	err = u.UnmarshalString(zeroString2)
	if err != nil {
		t.Fatal(err)
	}
	if !u.IsZero() {
		t.Fatal("Incorrect value (2)")
	}

	tooLargeString := "00000000000000000000000000000000000000000000000000000000000000000" // 65 chars
	err = u.UnmarshalString(tooLargeString)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	emptyString := ""
	err = u.UnmarshalString(emptyString)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	string25519 := "7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed"
	bytes25519, err := hex.DecodeString(string25519)
	if err != nil {
		t.Fatal(err)
	}
	uTrue := &Uint256{}
	err = uTrue.UnmarshalBinary(bytes25519)
	if err != nil {
		t.Fatal(err)
	}
	err = u.UnmarshalString(string25519)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("uint256s do not match")
	}

	// Bad string
	badString := "z123"
	err = u.UnmarshalString(badString)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
}

func TestUint256String(t *testing.T) {
	obj := &testObj{}
	str := obj.Value.String()
	if str != "" {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	str = u.String()
	if str != "" {
		t.Fatal("Should have raised error (1)")
	}

	u.SetZero()
	str = u.String()
	strTrue := "0"
	if str != strTrue {
		t.Fatal("Error in String (1)")
	}

	string25519 := "7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed"
	err := u.UnmarshalString(string25519)
	if err != nil {
		t.Fatal(err)
	}
	str = u.String()
	if str != string25519 {
		t.Fatal("Error in String (2)")
	}
}

func TestUint256FromToUint32Array(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.ToUint32Array()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	_, err = u.ToUint32Array()
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}
	byteSlice := make([]byte, 32)
	u32Array := [8]uint32{}

	byteSlice[0] = 0
	byteSlice[1] = 1
	byteSlice[2] = 2
	byteSlice[3] = 3
	u32Array[7], err = utils.UnmarshalUint32(byteSlice[0:4])
	if err != nil {
		t.Fatal(err)
	}

	byteSlice[4] = 4
	byteSlice[5] = 5
	byteSlice[6] = 6
	byteSlice[7] = 7
	u32Array[6], err = utils.UnmarshalUint32(byteSlice[4:8])
	if err != nil {
		t.Fatal(err)
	}

	byteSlice[8] = 8
	byteSlice[9] = 9
	byteSlice[10] = 10
	byteSlice[11] = 11
	u32Array[5], err = utils.UnmarshalUint32(byteSlice[8:12])
	if err != nil {
		t.Fatal(err)
	}

	byteSlice[12] = 12
	byteSlice[13] = 13
	byteSlice[14] = 14
	byteSlice[15] = 15
	u32Array[4], err = utils.UnmarshalUint32(byteSlice[12:16])
	if err != nil {
		t.Fatal(err)
	}

	byteSlice[16] = 16
	byteSlice[17] = 17
	byteSlice[18] = 18
	byteSlice[19] = 19
	u32Array[3], err = utils.UnmarshalUint32(byteSlice[16:20])
	if err != nil {
		t.Fatal(err)
	}

	byteSlice[20] = 20
	byteSlice[21] = 21
	byteSlice[22] = 22
	byteSlice[23] = 23
	u32Array[2], err = utils.UnmarshalUint32(byteSlice[20:24])
	if err != nil {
		t.Fatal(err)
	}

	byteSlice[24] = 24
	byteSlice[25] = 25
	byteSlice[26] = 26
	byteSlice[27] = 27
	u32Array[1], err = utils.UnmarshalUint32(byteSlice[24:28])
	if err != nil {
		t.Fatal(err)
	}

	byteSlice[28] = 28
	byteSlice[29] = 29
	byteSlice[30] = 30
	byteSlice[31] = 31
	u32Array[0], err = utils.UnmarshalUint32(byteSlice[28:32])
	if err != nil {
		t.Fatal(err)
	}

	v := &Uint256{}
	err = v.FromUint32Array(u32Array)
	if err != nil {
		t.Fatal(err)
	}

	err = u.UnmarshalBinary(byteSlice)
	if err != nil {
		t.Fatal(err)
	}
	strUint256, err := u.MarshalString()
	if err != nil {
		t.Fatal(err)
	}

	b := new(big.Int).SetBytes(byteSlice)
	strBig := fmt.Sprintf("%064x", b)
	if strUint256 != strBig {
		t.Fatal("Should agree (1)")
	}

	vStr, err := v.MarshalString()
	if err != nil {
		t.Fatal(err)
	}
	if strUint256 != vStr {
		t.Fatal("Should agree (2)")
	}

	z := v.Clone()
	uArray, err := z.ToUint32Array()
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < len(uArray); i++ {
		if uArray[i] != u32Array[i] {
			t.Fatal("Arrays do not match")
		}
	}
}

func TestUint256Clone(t *testing.T) {
	data := make([]byte, 32)
	data[0] = 1
	data[31] = 1
	u := &Uint256{}
	err := u.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	z := u.Clone()
	if !z.Eq(u) {
		t.Fatal("Should be equal (1)")
	}
	zBytes, err := z.MarshalBinary()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, zBytes) {
		t.Fatal("Bytes are not equal!")
	}
	v := z.Clone()
	if !z.Eq(v) {
		t.Fatal("Should be equal (2)")
	}
	_, err = u.Add(z, One())
	if err != nil {
		t.Fatal(err)
	}
	// z is the Clone of u; we should have u != z at this point.
	// Changing u should not change z
	if z.Eq(u) {
		t.Fatal("Should not be equal (1)")
	}

	_, err = v.Add(z, One())
	if err != nil {
		t.Fatal(err)
	}
	// v is the Clone of z; we should have v != z at this point.
	// Changing v should not change z
	if z.Eq(v) {
		t.Fatal("Should not be equal (2)")
	}

	a := &Uint256{}
	b := a.Clone()
	if !a.IsZero() || !b.IsZero() {
		t.Fatal("Both should be zero")
	}
}

func TestUint256FromBigInt(t *testing.T) {
	zBig := new(big.Int).SetInt64(0)
	obj := &testObj{}
	_, err := obj.Value.FromBigInt(zBig)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	_, err = u.FromBigInt(nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	buf := make([]byte, 33)
	buf[0] = 1
	buf[32] = 1
	bInt := new(big.Int).SetBytes(buf)
	_, err = u.FromBigInt(bInt)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}

	v, err := u.FromBigInt(zBig)
	if err != nil {
		t.Fatal(err)
	}
	if !v.Eq(u) {
		t.Fatal("Uint256s are not equal (1)")
	}
	uString, err := u.MarshalString()
	if err != nil {
		t.Fatal(err)
	}
	strBig := fmt.Sprintf("%064x", zBig)
	if uString != strBig {
		t.Fatal("Should agree")
	}

	string25519 := "7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed"
	bytes25519, err := hex.DecodeString(string25519)
	if err != nil {
		t.Fatal(err)
	}
	big25519 := new(big.Int).SetBytes(bytes25519)
	v, err = u.FromBigInt(big25519)
	if err != nil {
		t.Fatal(err)
	}
	if !v.Eq(u) {
		t.Fatal("Uint256s are not equal (2)")
	}
}

func TestUint256ToBigInt(t *testing.T) {
	u := &Uint256{}
	_, err := u.ToBigInt()
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u.SetZero()
	b, err := u.ToBigInt()
	if err != nil {
		t.Fatal(err)
	}
	if b.Sign() != 0 {
		t.Fatal("b should be 0")
	}

	bigOne := new(big.Int).SetUint64(1)
	u.SetOne()
	b, err = u.ToBigInt()
	if err != nil {
		t.Fatal(err)
	}
	if b.Cmp(bigOne) != 0 {
		t.Fatal("b should be 1")
	}

	big25519 := new(big.Int).SetUint64(25519)
	_, err = u.FromUint64(25519)
	if err != nil {
		t.Fatal(err)
	}
	b, err = u.ToBigInt()
	if err != nil {
		t.Fatal(err)
	}
	if b.Cmp(big25519) != 0 {
		t.Fatal("b should be 25519")
	}
}

func TestUint256FromUint64(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.FromUint64(0)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	u := &Uint256{}
	val := uint64(constants.MaxUint32)
	v, err := u.FromUint64(val)
	if err != nil {
		t.Fatal(err)
	}
	if !v.Eq(u) {
		t.Fatal("Uint256s are not equal (1)")
	}
	val = uint64(12345678901234567890)
	v, err = u.FromUint64(val)
	if err != nil {
		t.Fatal(err)
	}
	if !v.Eq(u) {
		t.Fatal("Uint256s are not equal (2)")
	}
}

func TestUint256ToUint64(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.ToUint64()
	if err == nil {
		t.Fatal("Should have raised error")
	}

	u := &Uint256{}
	z := uint64(0)
	ret0, err := u.ToUint64()
	if err != nil {
		t.Fatal(err)
	}
	if ret0 != z {
		t.Fatal("Returned incorrect value for 0")
	}
	val := constants.MaxUint64
	_, err = u.FromUint64(val)
	if err != nil {
		t.Fatal(err)
	}
	ret, err := u.ToUint64()
	if err != nil {
		t.Fatal(err)
	}
	if ret != val {
		t.Fatal("Returned incorrect value")
	}

	// Cause overflow from uint64 by adding 1
	_, err = u.Add(u, One())
	if err != nil {
		t.Fatal(err)
	}
	_, err = u.ToUint64()
	if err == nil {
		t.Fatal("Should have raised overflow error (1)")
	}

	data := make([]byte, 32)
	data[0] = 1
	data[31] = 1
	err = u.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = u.ToUint64()
	if err == nil {
		t.Fatal("Should have raised overflow error (2)")
	}
}

func TestUint256ToUint32(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.ToUint32()
	if err == nil {
		t.Fatal("Should have raised error")
	}

	u := &Uint256{}
	z := uint32(0)
	ret0, err := u.ToUint32()
	if err != nil {
		t.Fatal(err)
	}
	if ret0 != z {
		t.Fatal("Returned incorrect value for 0")
	}
	val := constants.MaxUint32
	_, err = u.FromUint64(uint64(val))
	if err != nil {
		t.Fatal(err)
	}
	ret, err := u.ToUint32()
	if err != nil {
		t.Fatal(err)
	}
	if ret != val {
		t.Fatal("Returned incorrect value")
	}

	// Cause overflow from uint64
	_, err = u.FromUint64(uint64(val) + 1)
	if err != nil {
		t.Fatal(err)
	}
	_, err = u.ToUint32()
	if err == nil {
		t.Fatal("Should have raised overflow error (1)")
	}

	data := make([]byte, 32)
	data[0] = 1
	data[31] = 1
	err = u.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = u.ToUint32()
	if err == nil {
		t.Fatal("Should have raised overflow error (2)")
	}
}

func TestUint256AddNils(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.Add(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	_, err = u.Add(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	a := &Uint256{}
	_, err = u.Add(a, nil)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	a.SetOne()

	b := &Uint256{}
	_, err = u.Add(a, b)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
	b.SetOne()

	u2, err := u.Add(a, b)
	if err != nil {
		t.Fatal(err)
	}
	two := Two()
	if !u.Eq(two) {
		t.Fatal("Incorrect value (1)")
	}
	if !u2.Eq(two) {
		t.Fatal("Incorrect value (2)")
	}
}

func TestUint256AddStandard(t *testing.T) {
	// Standard Test: save properly and do not overwrite args
	a := uint64(65537)
	b := uint64(25519)
	c := a + b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	z := new(Uint256)
	u, err := z.Add(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Add gives incorrect results (1)")
	}
	if !z.Eq(uTrue) {
		t.Fatal("Add gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Add changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Add changes input value 2")
	}
}

func TestUint256AddOverwrite1(t *testing.T) {
	// Standard Test: Argument reuse 1
	a := uint64(65537)
	b := uint64(25519)
	c := a + b
	x := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	_, err := x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	uTrue, err := new(Uint256).FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	u, err := x.Add(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Add gives incorrect results (1)")
	}
	if !x.Eq(uTrue) {
		t.Fatal("Add gives incorrect results (2)")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Add changes input value 2")
	}
}

func TestUint256AddOverwrite2(t *testing.T) {
	// Standard Test: Argument reuse 2
	a := uint64(65537)
	b := uint64(25519)
	c := a + b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	_, err := x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	uTrue, err := new(Uint256).FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	u, err := y.Add(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Add gives incorrect results (3)")
	}
	if !y.Eq(uTrue) {
		t.Fatal("Add gives incorrect results (4)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Add changes input value 1")
	}
}

func TestUint256AddOverflow(t *testing.T) {
	data := make([]byte, 32)
	data[0] = 255
	o1 := &Uint256{}
	err := o1.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	o2 := &Uint256{}
	err = o2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = new(Uint256).Add(o1, o2)
	if err == nil {
		t.Fatal("Add should raise error for overflow")
	}
}

func TestUint256AddModNils(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.AddMod(nil, nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	_, err = u.AddMod(nil, nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	a := &Uint256{}
	_, err = u.AddMod(a, nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	a.SetOne()

	b := &Uint256{}
	_, err = u.AddMod(a, b, nil)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
	b.SetOne()

	m := &Uint256{}
	_, err = u.AddMod(a, b, m)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}
	_, err = m.FromUint64(5)
	if err != nil {
		t.Fatal(err)
	}

	two := Two()
	res, err := u.AddMod(a, b, m)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(two) {
		t.Fatal("Incorrect result in AddMod (1)")
	}
	if !res.Eq(two) {
		t.Fatal("Incorrect result in AddMod (2)")
	}
}

func TestUint256AddModStandard(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1234)
	c := (a + b) % m
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	r := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := r.AddMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (1)")
	}
	if !r.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Add changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Add changes input value 2")
	}
	if !z.Eq(zBefore) {
		t.Fatal("Add changes input value 3")
	}
}

func TestUint256AddMod0(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(0)
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	r := &Uint256{}
	_, err := x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.AddMod(x, y, z)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Add changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Add changes input value 2")
	}
	if !z.Eq(zBefore) {
		t.Fatal("Add changes input value 3")
	}
}

func TestUint256AddMod1(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1)
	c := (a + b) % m
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	r := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := r.AddMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (1)")
	}
	if !r.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Add changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Add changes input value 2")
	}
	if !z.Eq(zBefore) {
		t.Fatal("Add changes input value 3")
	}
}

func TestUint256AddModOverwrite1(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1234)
	c := (a + b) % m
	x := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := x.AddMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (1)")
	}
	if !x.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (2)")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Add changes input value 2")
	}
	if !z.Eq(zBefore) {
		t.Fatal("Add changes input value 3")
	}
}

func TestUint256AddModOverwrite2(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1234)
	c := (a + b) % m
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := y.AddMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (1)")
	}
	if !y.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Add changes input value 1")
	}
	if !z.Eq(zBefore) {
		t.Fatal("Add changes input value 3")
	}
}

func TestUint256AddModOverwrite3(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1234)
	c := (a + b) % m
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := z.AddMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (1)")
	}
	if !z.Eq(uTrue) {
		t.Fatal("AddMod gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("AddMod changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("AddMod changes input value 2")
	}
}

func TestUint256SubNils(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.Sub(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	u := &Uint256{}
	_, err = u.Sub(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	a := &Uint256{}
	_, err = u.Sub(a, nil)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	a.SetOne()

	b := &Uint256{}
	_, err = u.Sub(a, b)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
	b.SetOne()

	u2, err := u.Sub(a, b)
	if err != nil {
		t.Fatal(err)
	}
	zero := Zero()
	if !u.Eq(zero) {
		t.Fatal("Incorrect value (1)")
	}
	if !u2.Eq(zero) {
		t.Fatal("Incorrect value (2)")
	}
}

func TestUint256SubStandard(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a - b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := z.Sub(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Sub gives incorrect results")
	}
	if !z.Eq(uTrue) {
		t.Fatal("Sub gives incorrect results")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Sub changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Sub changes input value 2")
	}
}

func TestUint256SubOverwrite1(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a - b
	x := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := x.Sub(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Sub gives incorrect results")
	}
	if !x.Eq(uTrue) {
		t.Fatal("Sub gives incorrect results")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Sub changes input value 2")
	}
}

func TestUint256SubOverwrite2(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a - b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := y.Sub(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Sub gives incorrect results")
	}
	if !y.Eq(uTrue) {
		t.Fatal("Sub gives incorrect results")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Sub changes input value 1")
	}
}

func TestUint256SubOverflow(t *testing.T) {
	o1 := &Uint256{}
	_, err := o1.FromUint64(1)
	if err != nil {
		t.Fatal(err)
	}
	o2 := &Uint256{}
	_, err = o2.FromUint64(3)
	if err != nil {
		t.Fatal(err)
	}
	_, err = new(Uint256).Sub(o1, o2)
	if err == nil {
		t.Fatal("Sub should raise error for overflow")
	}
}

func TestUint256MulNils(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.Mul(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (1)")
	}

	u := &Uint256{}
	_, err = u.Mul(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	a := &Uint256{}
	_, err = u.Mul(a, nil)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	a.SetOne()

	b := &Uint256{}
	_, err = u.Mul(a, b)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
	b.SetOne()

	one := One()
	r, err := u.Mul(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Eq(one) {
		t.Fatal("Mul returned incorrect result (1)")
	}
	if !u.Eq(one) {
		t.Fatal("Mul returned incorrect result (2)")
	}
}

func TestUint256MulStandard(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a * b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := z.Mul(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Mul gives incorrect results (1)")
	}
	if !z.Eq(uTrue) {
		t.Fatal("Mul gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Mul changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Mul changes input value 2")
	}
}

func TestUint256MulOverwrite1(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a * b
	x := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := x.Mul(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Mul gives incorrect results (1)")
	}
	if !x.Eq(uTrue) {
		t.Fatal("Mul gives incorrect results (2)")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Mul changes input value 2")
	}
}

func TestUint256MulOverwrite2(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a * b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := y.Mul(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Mul gives incorrect results (1)")
	}
	if !y.Eq(uTrue) {
		t.Fatal("Mul gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Mul changes input value 1")
	}
}

func TestUint256MulOverflow(t *testing.T) {
	data := make([]byte, 32)
	data[0] = 255
	o1 := &Uint256{}
	err := o1.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	o2 := &Uint256{}
	err = o2.UnmarshalBinary(data)
	if err != nil {
		t.Fatal(err)
	}
	_, err = new(Uint256).Mul(o1, o2)
	if err == nil {
		t.Fatal("Mul hould raise error for overflow")
	}
}

func TestUint256MulModNils(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.MulMod(nil, nil, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	u := &Uint256{}
	_, err = u.MulMod(nil, nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	a := &Uint256{}
	_, err = u.MulMod(a, nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	a.SetOne()

	b := &Uint256{}
	_, err = u.MulMod(a, b, nil)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
	b.SetOne()

	m := &Uint256{}
	_, err = u.MulMod(a, b, m)
	if err == nil {
		t.Fatal("Should have raised error (5)")
	}
	_, err = m.FromUint64(5)
	if err != nil {
		t.Fatal(err)
	}

	one := One()
	r, err := u.MulMod(a, b, m)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Eq(one) {
		t.Fatal("MulMod returned incorrect result (1)")
	}
	if !u.Eq(one) {
		t.Fatal("MulMod returned incorrect result (2)")
	}
}

func TestUint256MulModStandard(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1234)
	c := (a * b) % m
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	r := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := r.MulMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !r.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !x.Eq(xBefore) {
		t.Fatal("MulMod changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("MulMod changes input value 2")
	}
	if !z.Eq(zBefore) {
		t.Fatal("MulMod changes input value 3")
	}
}

func TestUint256MulMod0(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(0)
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	r := &Uint256{}
	_, err := x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = r.MulMod(x, y, z)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !x.Eq(xBefore) {
		t.Fatal("MulMod changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("MulMod changes input value 2")
	}
	if !z.Eq(zBefore) {
		t.Fatal("MulMod changes input value 3")
	}
}

func TestUint256MulMod1(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1)
	c := (a * b) % m
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	r := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := r.MulMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !r.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !x.Eq(xBefore) {
		t.Fatal("MulMod changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("MulMod changes input value 2")
	}
	if !z.Eq(zBefore) {
		t.Fatal("MulMod changes input value 3")
	}
}

func TestUint256MulModOverwrite1(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1234)
	c := (a * b) % m
	x := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := x.MulMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !x.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !y.Eq(yBefore) {
		t.Fatal("MulMod changes input value 2")
	}
	if !z.Eq(zBefore) {
		t.Fatal("MulMod changes input value 3")
	}
}

func TestUint256MulModOverwrite2(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1234)
	c := (a * b) % m
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	z := &Uint256{}
	zBefore := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	_, err = zBefore.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := y.MulMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !y.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !x.Eq(xBefore) {
		t.Fatal("MulMod changes input value 1")
	}
	if !z.Eq(zBefore) {
		t.Fatal("MulMod changes input value 3")
	}
}

func TestUint256MulModOverwrite3(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	m := uint64(1234)
	c := (a * b) % m
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.FromUint64(m)
	if err != nil {
		t.Fatal(err)
	}
	u, err := z.MulMod(x, y, z)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !z.Eq(uTrue) {
		t.Fatal("MulMod gives incorrect results")
	}
	if !x.Eq(xBefore) {
		t.Fatal("MulMod changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("MulMod changes input value 2")
	}
}

func TestUint256DivNils(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.Div(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	u := &Uint256{}
	_, err = u.Div(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	a := &Uint256{}
	_, err = u.Div(a, nil)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	a.SetOne()

	b := &Uint256{}
	_, err = u.Div(a, b)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
	b.SetOne()

	one := One()
	r, err := u.Div(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Eq(one) {
		t.Fatal("Div returns incorrect result (1)")
	}
	if !u.Eq(one) {
		t.Fatal("Div returns incorrect result (2)")
	}
}

func TestUint256DivStandard(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a / b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := z.Div(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Div gives incorrect results (1)")
	}
	if !z.Eq(uTrue) {
		t.Fatal("Div gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Div changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Div changes input value 2")
	}
}

func TestUint256Div0(t *testing.T) {
	a := uint64(65537)
	b := uint64(0)
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	_, err := x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.Div(x, y)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Div changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Div changes input value 2")
	}
}

func TestUint256DivOverwrite1(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a / b
	x := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := x.Div(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Div gives incorrect results (1)")
	}
	if !x.Eq(uTrue) {
		t.Fatal("Div gives incorrect results (2)")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Div changes input value 2")
	}
}

func TestUint256DivOverwrite2(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a / b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := y.Div(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Div gives incorrect results (1)")
	}
	if !y.Eq(uTrue) {
		t.Fatal("Div gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Div changes input value 1")
	}
}

func TestUint256ModNils(t *testing.T) {
	obj := &testObj{}
	_, err := obj.Value.Mod(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error")
	}

	u := &Uint256{}
	_, err = u.Mod(nil, nil)
	if err == nil {
		t.Fatal("Should have raised error (2)")
	}

	a := &Uint256{}
	_, err = u.Mod(a, nil)
	if err == nil {
		t.Fatal("Should have raised error (3)")
	}
	a.SetOne()

	m := &Uint256{}
	_, err = u.Mod(a, m)
	if err == nil {
		t.Fatal("Should have raised error (4)")
	}
	_, err = m.FromUint64(5)
	if err != nil {
		t.Fatal(err)
	}

	one := One()
	r, err := u.Mod(a, m)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Eq(one) {
		t.Fatal("Mod returns incorrect result (1)")
	}
	if !u.Eq(one) {
		t.Fatal("Mod returns incorrect result (2)")
	}
}

func TestUint256ModStandard(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a % b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := z.Mod(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Mod gives incorrect results (1)")
	}
	if !z.Eq(uTrue) {
		t.Fatal("Mod gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Mod changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Mod changes input value 2")
	}
}

func TestUint256Mod0(t *testing.T) {
	a := uint64(65537)
	b := uint64(0)
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	z := &Uint256{}
	_, err := x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = z.Mod(x, y)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Mod changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Mod changes input value 2")
	}
}

func TestUint256ModStandard2(t *testing.T) {
	// 2^255
	xSlice, err := hex.DecodeString("8000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Fatal(err)
	}
	x := &Uint256{}
	err = x.UnmarshalBinary(xSlice)
	if err != nil {
		t.Fatal(err)
	}
	xBefore := &Uint256{}
	err = xBefore.UnmarshalBinary(xSlice)
	if err != nil {
		t.Fatal(err)
	}
	// 2^255 - 19
	ySlice, err := hex.DecodeString("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffed")
	if err != nil {
		t.Fatal(err)
	}
	y := &Uint256{}
	err = y.UnmarshalBinary(ySlice)
	if err != nil {
		t.Fatal(err)
	}
	yBefore := &Uint256{}
	err = yBefore.UnmarshalBinary(ySlice)
	if err != nil {
		t.Fatal(err)
	}
	retTrue, err := new(Uint256).FromUint64(19)
	if err != nil {
		t.Fatal(err)
	}
	z := &Uint256{}
	ret, err := z.Mod(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !ret.Eq(retTrue) {
		t.Fatal("Mod gives incorrect results (1)")
	}
	if !z.Eq(retTrue) {
		t.Fatal("Mod gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Mod changes input value 1")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Mod changes input value 2")
	}
}

func TestUint256ModOverwrite1(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a % b
	x := &Uint256{}
	y := &Uint256{}
	yBefore := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	_, err = yBefore.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := x.Mod(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Mod gives incorrect results (1)")
	}
	if !x.Eq(uTrue) {
		t.Fatal("Mod gives incorrect results (2)")
	}
	if !y.Eq(yBefore) {
		t.Fatal("Mod changes input value 2")
	}
}

func TestUint256ModOverwrite2(t *testing.T) {
	a := uint64(65537)
	b := uint64(25519)
	c := a % b
	x := &Uint256{}
	xBefore := &Uint256{}
	y := &Uint256{}
	uTrue := &Uint256{}
	_, err := uTrue.FromUint64(c)
	if err != nil {
		t.Fatal(err)
	}
	_, err = x.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = xBefore.FromUint64(a)
	if err != nil {
		t.Fatal(err)
	}
	_, err = y.FromUint64(b)
	if err != nil {
		t.Fatal(err)
	}
	u, err := y.Mod(x, y)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Mod gives incorrect results (1)")
	}
	if !y.Eq(uTrue) {
		t.Fatal("Mod gives incorrect results (2)")
	}
	if !x.Eq(xBefore) {
		t.Fatal("Mod changes input value 1")
	}
}

func TestUint256GreaterThan(t *testing.T) {
	u := Zero()
	v := One()
	if !v.Gt(u) {
		t.Fatal("Gt fails")
	}
}

func TestUint256LessThan(t *testing.T) {
	u := Zero()
	v := One()
	if !u.Lt(v) {
		t.Fatal("Lt fails")
	}
}

func TestUint256Equal(t *testing.T) {
	u := Zero()
	v := One()
	x := Zero()
	if !u.Eq(u) {
		t.Fatal("Eq fails (1)")
	}
	if u.Eq(v) {
		t.Fatal("Eq fails (2)")
	}
	if !u.Eq(x) {
		t.Fatal("Eq fails (3)")
	}
}

func TestUint256Compare(t *testing.T) {
	u := Zero()
	v := One()
	if u.Cmp(u) != 0 {
		t.Fatal("Cmp fails (1)")
	}
	if u.Cmp(v) != -1 {
		t.Fatal("Cmp fails (2)")
	}
	if v.Cmp(u) != 1 {
		t.Fatal("Cmp fails (3)")
	}
}

func TestUint256GreaterThanOrEqual(t *testing.T) {
	u := Zero()
	v := One()
	x := Zero()
	if !u.Gte(u) {
		t.Fatal("Gte fails (1)")
	}
	if u.Gte(v) {
		t.Fatal("Gte fails (2)")
	}
	if !u.Gte(x) {
		t.Fatal("Gte fails (3)")
	}
}

func TestUint256LessThanOrEqual(t *testing.T) {
	u := Zero()
	v := One()
	x := Zero()
	if !u.Lte(u) {
		t.Fatal("Lte fails (1)")
	}
	if v.Lte(u) {
		t.Fatal("Lte fails (2)")
	}
	if !u.Lte(x) {
		t.Fatal("Lte fails (3)")
	}
}

func TestUint256DSPIMinDeposit(t *testing.T) {
	u := DSPIMinDeposit()
	uTrue, err := new(Uint256).FromUint64(uint64(constants.DSPIMinDeposit))
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("DSPIMinDeposit() fails")
	}
}

func TestUint256BaseDatasizeConst(t *testing.T) {
	u := BaseDatasizeConst()
	uTrue, err := new(Uint256).FromUint64(uint64(constants.BaseDatasizeConst))
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("BaseDatasizeConst() fails")
	}
}

func TestUint256Zero(t *testing.T) {
	u := Zero()
	uTrue := new(Uint256).SetZero()
	if !u.Eq(uTrue) {
		t.Fatal("Zero() fails")
	}
}

func TestUint256IsZero(t *testing.T) {
	u := Zero()
	if !u.IsZero() {
		t.Fatal("IsZero() should return true")
	}

	u = One()
	if u.IsZero() {
		t.Fatal("IsZero() should return false")
	}
}

func TestUint256One(t *testing.T) {
	u := One()
	uTrue := new(Uint256).SetOne()
	if !u.Eq(uTrue) {
		t.Fatal("One() fails")
	}
}

func TestUint256Two(t *testing.T) {
	u := Two()
	uTrue := new(Uint256).SetOne()
	uTrue, err := new(Uint256).Add(uTrue, uTrue)
	if err != nil {
		t.Fatal(err)
	}
	if !u.Eq(uTrue) {
		t.Fatal("Two() fails")
	}
}
