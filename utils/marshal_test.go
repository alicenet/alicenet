package utils

import (
	"bytes"

	"github.com/MadBase/MadNet/constants"

	"testing"
)

func TestMarshalUint16(t *testing.T) {
	z := uint16(0)
	zBytesTrue := make([]byte, 2)
	zBytes := MarshalUint16(z)
	if !bytes.Equal(zBytes, zBytesTrue) {
		t.Fatal("MarshalUint16 fail: 0")
	}
	zUn, err := UnmarshalUint16(zBytes)
	if err != nil {
		t.Fatal(err)
	}
	if z != zUn {
		t.Fatal("UnmarshalUint16 fail: 0")
	}

	maxUint16 := uint16(65535)
	maxUint16BytesTrue := make([]byte, 2)
	for i := 0; i < len(maxUint16BytesTrue); i++ {
		maxUint16BytesTrue[i] = 255
	}
	maxUint16Bytes := MarshalUint16(maxUint16)
	if !bytes.Equal(maxUint16Bytes, maxUint16BytesTrue) {
		t.Fatal("MarshalUint16 fail: MaxUint16")
	}
	maxUint16Un, err := UnmarshalUint16(maxUint16Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if maxUint16 != maxUint16Un {
		t.Fatal("UnmarshalUint16 fail: MaxUint16")
	}

	u16v257 := uint16(257)
	u16v257BytesTrue := make([]byte, 2)
	u16v257BytesTrue[1] = 1
	u16v257BytesTrue[0] = 1
	u16v257Bytes := MarshalUint16(u16v257)
	if !bytes.Equal(u16v257Bytes, u16v257BytesTrue) {
		t.Fatal("MarshalUint16 fail: 257")
	}
	u16v257Un, err := UnmarshalUint16(u16v257Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if u16v257 != u16v257Un {
		t.Fatal("UnmarshalUint16 fail: 257")
	}

	badBytes := make([]byte, 3)
	_, err = UnmarshalUint16(badBytes)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestMarshalUint32(t *testing.T) {
	z := uint32(0)
	zBytesTrue := make([]byte, 4)
	zBytes := MarshalUint32(z)
	if !bytes.Equal(zBytes, zBytesTrue) {
		t.Fatal("MarshalUint32 fail: 0")
	}
	zUn, err := UnmarshalUint32(zBytes)
	if err != nil {
		t.Fatal(err)
	}
	if z != zUn {
		t.Fatal("UnmarshalUint32 fail: 0")

	}

	maxUint32 := constants.MaxUint32
	maxUint32BytesTrue := make([]byte, 4)
	for i := 0; i < len(maxUint32BytesTrue); i++ {
		maxUint32BytesTrue[i] = 255
	}
	maxUint32Bytes := MarshalUint32(maxUint32)
	if !bytes.Equal(maxUint32Bytes, maxUint32BytesTrue) {
		t.Fatal("MarshalUint32 fail: MaxUint32")
	}
	maxUint32Un, err := UnmarshalUint32(maxUint32Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if maxUint32 != maxUint32Un {
		t.Fatal("UnmarshalUint32 fail: MaxUint32")
	}

	u32v257 := uint32(257)
	u32v257BytesTrue := make([]byte, 4)
	u32v257BytesTrue[3] = 1
	u32v257BytesTrue[2] = 1
	u32v257Bytes := MarshalUint32(u32v257)
	if !bytes.Equal(u32v257Bytes, u32v257BytesTrue) {
		t.Fatal("MarshalUint32 fail: 257")
	}
	u32v257Un, err := UnmarshalUint32(u32v257Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if u32v257 != u32v257Un {
		t.Fatal("UnmarshalUint32 fail: 257")
	}

	u32v65537 := uint32(65537)
	u32v65537BytesTrue := make([]byte, 4)
	u32v65537BytesTrue[3] = 1
	u32v65537BytesTrue[1] = 1
	u32v65537Bytes := MarshalUint32(u32v65537)
	if !bytes.Equal(u32v65537Bytes, u32v65537BytesTrue) {
		t.Fatal("MarshalUint32 fail: 65537")
	}
	u32v65537Un, err := UnmarshalUint32(u32v65537Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if u32v65537 != u32v65537Un {
		t.Fatal("UnmarshalUint32 fail: 65537")
	}

	u32v16777217 := uint32(16777217)
	u32v16777217BytesTrue := make([]byte, 4)
	u32v16777217BytesTrue[3] = 1
	u32v16777217BytesTrue[0] = 1
	u32v16777217Bytes := MarshalUint32(u32v16777217)
	if !bytes.Equal(u32v16777217Bytes, u32v16777217BytesTrue) {
		t.Fatal("MarshalUint32 fail: 16777217")
	}
	u32v16777217Un, err := UnmarshalUint32(u32v16777217Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if u32v16777217 != u32v16777217Un {
		t.Fatal("UnmarshalUint32 fail: 16777217")
	}

	badBytes := make([]byte, 5)
	_, err = UnmarshalUint32(badBytes)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestMarshalUint64(t *testing.T) {
	z := uint64(0)
	zBytesTrue := make([]byte, 8)
	zBytes := MarshalUint64(z)
	if !bytes.Equal(zBytes, zBytesTrue) {
		t.Fatal("MarshalUint64 fail: 0")
	}
	zUn, err := UnmarshalUint64(zBytes)
	if err != nil {
		t.Fatal(err)
	}
	if z != zUn {
		t.Fatal("UnmarshalUint64 fail: 0")
	}

	p1 := uint64(1)
	p1Bytes := MarshalUint64(p1)
	p1Un, err := UnmarshalUint64(p1Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if p1 != p1Un {
		t.Fatal("UnmarshalUint64 fail: 1")
	}

	n1 := constants.MaxUint64
	n1Bytes := MarshalUint64(n1)
	n1Un, err := UnmarshalUint64(n1Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if n1 != n1Un {
		t.Fatal("UnmarshalUint64 fail: MaxUint64")
	}

	badBytes := make([]byte, 9)
	_, err = UnmarshalUint64(badBytes)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}

func TestMarshalInt64(t *testing.T) {
	z := int64(0)
	zBytesTrue := make([]byte, 8)
	zBytes := MarshalInt64(z)
	if !bytes.Equal(zBytes, zBytesTrue) {
		t.Fatal("MarshalInt64 fail: 0")
	}
	zUn, err := UnmarshalInt64(zBytes)
	if err != nil {
		t.Fatal(err)
	}
	if z != zUn {
		t.Fatal("UnmarshalInt64 fail: 0")
	}

	p1 := int64(1)
	p1Bytes := MarshalInt64(p1)
	p1Un, err := UnmarshalInt64(p1Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if p1 != p1Un {
		t.Fatal("UnmarshalInt64 fail: 1")
	}

	n1 := int64(-1)
	n1Bytes := MarshalInt64(n1)
	n1Un, err := UnmarshalInt64(n1Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if n1 != n1Un {
		t.Fatal("UnmarshalInt64 fail: -1")
	}

	badBytes := make([]byte, 9)
	_, err = UnmarshalInt64(badBytes)
	if err == nil {
		t.Fatal("Should have raised error")
	}
}
