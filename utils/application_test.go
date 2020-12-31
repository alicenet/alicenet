package utils

import (
	"errors"

	"testing"
)

type testObj struct {
	data []byte
}

func (to *testObj) MarshalBinary() ([]byte, error) {
	if len(to.data) == 0 {
		return nil, errors.New("Empty Object")
	}
	return CopySlice(to.data), nil
}

func (to *testObj) UnmarshalBinary(data []byte) error {
	if len(data) == 0 {
		return errors.New("Nil Data")
	}
	to.data = CopySlice(to.data)
	return nil
}

func TestGetObjSize(t *testing.T) {
	to := &testObj{}
	_, err := GetObjSize(to)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	dataSizeTrue := 32
	data := make([]byte, dataSizeTrue)
	data[0] = 1
	data[dataSizeTrue-1] = 1
	to.data = CopySlice(data)
	dataSize, err := GetObjSize(to)
	if err != nil {
		t.Fatal(err)
	}
	if dataSize != dataSizeTrue {
		t.Fatal("dataSize does not match true value")
	}
}
