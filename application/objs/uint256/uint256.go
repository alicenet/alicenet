package uint256

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/utils"

	"github.com/holiman/uint256"
)

// TODO: clone a little bit and confirm no nil pointer deferences

// Uint256 is an unsigned 256-bit integer
type Uint256 struct {
	val *uint256.Int
}

// MarshalBinary marshals Uint256 to a byte slice
func (u *Uint256) MarshalBinary() ([]byte, error) {
	if u == nil || u.val == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.MarshalBinary: nil object")
	}
	buf := u.val.Bytes32()
	return buf[:], nil
}

// UnmarshalBinary unmarshals a byte slice to a Uint256 object
func (u *Uint256) UnmarshalBinary(data []byte) error {
	if u == nil {
		return errorz.ErrInvalid{}.New("Error in Uint256.UnmarshalBinary: nil object")
	}
	if len(data) == 0 {
		return errorz.ErrInvalid{}.New("Error in Uint256.UnmarshalBinary: empty slice")
	}
	if len(data) > 32 {
		return errorz.ErrInvalid{}.New("Error in Uint256.UnmarshalBinary: data slice too long")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	data32 := utils.ForceSliceToLength(data, 32)
	u.val.SetBytes(data32)
	return nil
}

// MarshalString returns 64 hex-encoded string of Uint256 object
func (u *Uint256) MarshalString() (string, error) {
	b, err := u.MarshalBinary()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// UnmarshalString marshals hex-encoded string to Uint256 object
func (u *Uint256) UnmarshalString(s string) error {
	if len(s) > 64 {
		return errorz.ErrInvalid{}.New("Error in Uint256.UnmarshalString: string too long")
	}
	if len(s) == 0 {
		return errorz.ErrInvalid{}.New("Error in Uint256.UnmarshalString: empty string")
	}
	padded := fmt.Sprintf("%064v", s)
	b, err := hex.DecodeString(padded)
	if err != nil {
		return errorz.ErrInvalid{}.New("Error in Uint256.UnmarshalString: DecodeString error")
	}
	return u.UnmarshalBinary(b)
}

// String returns the result of MarshalString
func (u *Uint256) String() string {
	_, err := u.MarshalString()
	if err != nil {
		return ""
	}
	s := u.val.String()
	prefix := "0x"
	return strings.TrimPrefix(s, prefix)
}

// ToUint32Array converts Uint256 into an array of uint32 objects;
// note that the order is reversed, so that the lower-order uint32 objects
// have smaller indices.
func (u *Uint256) ToUint32Array() ([8]uint32, error) {
	b, err := u.MarshalBinary()
	if err != nil {
		return [8]uint32{}, err
	}
	z := [8]uint32{}
	z[7], _ = utils.UnmarshalUint32(b[0:4])
	z[6], _ = utils.UnmarshalUint32(b[4:8])
	z[5], _ = utils.UnmarshalUint32(b[8:12])
	z[4], _ = utils.UnmarshalUint32(b[12:16])
	z[3], _ = utils.UnmarshalUint32(b[16:20])
	z[2], _ = utils.UnmarshalUint32(b[20:24])
	z[1], _ = utils.UnmarshalUint32(b[24:28])
	z[0], _ = utils.UnmarshalUint32(b[28:32])
	return z, nil
}

// FromUint32Array takes in an array of uint32 objects and make a Uint256
func (u *Uint256) FromUint32Array(z [8]uint32) error {
	b := make([]byte, 32)
	copy(b[0:4], utils.MarshalUint32(z[7]))
	copy(b[4:8], utils.MarshalUint32(z[6]))
	copy(b[8:12], utils.MarshalUint32(z[5]))
	copy(b[12:16], utils.MarshalUint32(z[4]))
	copy(b[16:20], utils.MarshalUint32(z[3]))
	copy(b[20:24], utils.MarshalUint32(z[2]))
	copy(b[24:28], utils.MarshalUint32(z[1]))
	copy(b[28:32], utils.MarshalUint32(z[0]))
	return u.UnmarshalBinary(b)
}

// Clone returns a copy of u; will panic if u == nil
func (u *Uint256) Clone() *Uint256 {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	v := &Uint256{}
	v.val = &uint256.Int{}
	v.val = u.val.Clone()
	return v
}

// FromBigInt converts big.Int into Uint256
func (u *Uint256) FromBigInt(a *big.Int) (*Uint256, error) {
	if a == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.FromBigInt: nil arg")
	}
	buf := a.Bytes()
	if len(buf) > 32 {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.FromBigInt: data slice too long")
	}
	buf32 := utils.ForceSliceToLength(buf, 32)
	err := u.UnmarshalBinary(buf32)
	if err != nil {
		return nil, err
	}
	return u.Clone(), nil
}

// ToBigInt converts Uint256 into big.Int
func (u *Uint256) ToBigInt() (*big.Int, error) {
	buf, err := u.MarshalBinary()
	if err != nil {
		return nil, err
	}
	a := new(big.Int).SetBytes(buf)
	return a, nil
}

// FromUint64 converts a uint64 into Uint256
func (u *Uint256) FromUint64(a uint64) (*Uint256, error) {
	buf := utils.MarshalUint64(a)
	err := u.UnmarshalBinary(buf)
	if err != nil {
		return nil, err
	}
	return u.Clone(), nil
}

// ToUint64 returns the lowest 64 bits of u and an error if
// the operation overflows
func (u *Uint256) ToUint64() (uint64, error) {
	if u == nil {
		return uint64(0), errorz.ErrInvalid{}.New("Error in Uint256.ToUint64: not initialized")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	u64, overflowed := u.val.Uint64WithOverflow()
	if overflowed {
		return uint64(0), errorz.ErrInvalid{}.New("Error in Uint256.ToUint64: overflow")
	}
	return u64, nil
}

// ToUint32 returns the lowest 32 bits of u and an error if
// the operation overflows
func (u *Uint256) ToUint32() (uint32, error) {
	if u == nil {
		return uint32(0), errorz.ErrInvalid{}.New("Error in Uint256.ToUint32: not initialized")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	u64, overflowed := u.val.Uint64WithOverflow()
	if overflowed || u64 > uint64(constants.MaxUint32) {
		return uint32(0), errorz.ErrInvalid{}.New("Error in Uint256.ToUint32: overflow")
	}
	return uint32(u64), nil
}

// Add returns u == a + b and returns an error on overflow
func (u *Uint256) Add(a, b *Uint256) (*Uint256, error) {
	if u == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Add: not initialized")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	if a == nil || a.val == nil || b == nil || b.val == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Add: nil args")
	}
	z := u.Clone()
	_, overflowed := z.val.AddOverflow(a.val, b.val)
	if overflowed {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Add: overflow")
	}
	u.val = z.val.Clone()
	return z, nil
}

// AddMod returns u == a + b mod m
func (u *Uint256) AddMod(a, b, m *Uint256) (*Uint256, error) {
	if u == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.AddMod: not initialized")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	if a == nil || a.val == nil || b == nil || b.val == nil || m == nil || m.val == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.AddMod: nil args")
	}
	if m.Eq(Zero()) {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.AddMod: Cannot mod by 0")
	}
	u.val.AddMod(a.val, b.val, m.val)
	return u.Clone(), nil
}

// Sub returns u == a - b and returns an error on overflow
func (u *Uint256) Sub(a, b *Uint256) (*Uint256, error) {
	if u == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Sub: not initialized")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	if a == nil || a.val == nil || b == nil || b.val == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Sub: nil args")
	}
	z := u.Clone()
	_, overflowed := z.val.SubOverflow(a.val, b.val)
	if overflowed {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Sub: overflow")
	}
	u.val = z.val.Clone()
	return z, nil
}

// Mul returns u == a * b and returns an error on overflow
func (u *Uint256) Mul(a, b *Uint256) (*Uint256, error) {
	if u == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Mul: not initialized")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	if a == nil || a.val == nil || b == nil || b.val == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Mul: nil args")
	}
	z := u.Clone()
	_, overflowed := z.val.MulOverflow(a.val, b.val)
	if overflowed {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Mul: overflow")
	}
	u.val = z.val.Clone()
	return z, nil
}

// MulMod returns u == a * b mod m
func (u *Uint256) MulMod(a, b, m *Uint256) (*Uint256, error) {
	if u == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.MulMod: not initialized")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	if a == nil || a.val == nil || b == nil || b.val == nil || m == nil || m.val == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.MulMod: nil args")
	}
	if m.Eq(Zero()) {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.MulMod: Cannot mod by 0")
	}
	u.val.MulMod(a.val, b.val, m.val)
	return u.Clone(), nil
}

// Div returns u == a / b
func (u *Uint256) Div(a, b *Uint256) (*Uint256, error) {
	if u == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Div: not initialized")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	if a == nil || a.val == nil || b == nil || b.val == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Div: nil args")
	}
	if b.Eq(Zero()) {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Mod: Cannot divide by 0")
	}
	u.val.Div(a.val, b.val)
	return u.Clone(), nil
}

// Mod returns u == a mod m
func (u *Uint256) Mod(a, m *Uint256) (*Uint256, error) {
	if u == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Mod: not initialized")
	}
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	if a == nil || a.val == nil || m == nil || m.val == nil {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Mod: nil args")
	}
	if m.Eq(Zero()) {
		return nil, errorz.ErrInvalid{}.New("Error in Uint256.Mod: Cannot mod by 0")
	}
	u.val.Mod(a.val, m.val)
	return u.Clone(), nil
}

// Gt returns u > a
func (u *Uint256) Gt(a *Uint256) bool {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	return u.val.Gt(a.val)
}

// Gte returns u >= a
func (u *Uint256) Gte(a *Uint256) bool {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	return u.Gt(a) || u.Eq(a)
}

// Lt returns u < a
func (u *Uint256) Lt(a *Uint256) bool {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	return u.val.Lt(a.val)
}

// Lte returns u <= a
func (u *Uint256) Lte(a *Uint256) bool {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	return u.Lt(a) || u.Eq(a)
}

// Eq returns u == a
func (u *Uint256) Eq(a *Uint256) bool {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	return u.val.Eq(a.val)
}

// Cmp returns 1 if u > a, 0 if u == a, and -1 if u < a
func (u *Uint256) Cmp(a *Uint256) int {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	return u.val.Cmp(a.val)
}

// SetOne sets u to 1
func (u *Uint256) SetOne() *Uint256 {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	u.val.SetOne()
	return u.Clone()
}

// SetZero sets u to 0
func (u *Uint256) SetZero() *Uint256 {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	u.val.Clear()
	return u.Clone()
}

// IsZero determines if u is 0
func (u *Uint256) IsZero() bool {
	if u.val == nil {
		u.val = &uint256.Int{}
	}
	return u.val.IsZero()
}

// DSPIMinDeposit returns constants.DSPIMinDeposit as Uint256
func DSPIMinDeposit() *Uint256 {
	u, _ := new(Uint256).FromUint64(uint64(constants.DSPIMinDeposit))
	return u
}

// BaseDatasizeConst returns constants.BaseDatasizeConst as Uint256
func BaseDatasizeConst() *Uint256 {
	u, _ := new(Uint256).FromUint64(uint64(constants.BaseDatasizeConst))
	return u
}

// Zero returns the Uint256 zero
func Zero() *Uint256 {
	return new(Uint256).SetZero()
}

// One returns the Uint256 one
func One() *Uint256 {
	return new(Uint256).SetOne()
}

// Two returns the Uint256 two
func Two() *Uint256 {
	u := &Uint256{}
	_, _ = u.FromUint64(2)
	return u
}

// Max returns the maximum value
func Max() *Uint256 {
	const uint64max uint64 = 1<<64 - 1
	return &Uint256{val: &uint256.Int{uint64max, uint64max, uint64max, uint64max}}
}
