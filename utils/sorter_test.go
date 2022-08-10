package utils

import (
	"bytes"
	"testing"
)

func TestSorter(t *testing.T) {
	keyLength := 32
	key0 := make([]byte, keyLength) // 257
	key0[keyLength-1] = 1
	key0[keyLength-2] = 1
	key1 := make([]byte, keyLength) // 2
	key1[keyLength-1] = 2
	key2 := make([]byte, keyLength) // 255
	key2[keyLength-1] = 255
	key3 := make([]byte, keyLength) // 155
	key3[keyLength-1] = 155
	keys := make([][]byte, 4)
	keys[0] = CopySlice(key0)
	keys[1] = CopySlice(key1)
	keys[2] = CopySlice(key2)
	keys[3] = CopySlice(key3)
	// Correct sorting: 1, 3, 2, 0

	valueLength := 32
	value0 := make([]byte, valueLength)
	value0[valueLength-1] = 2
	value1 := make([]byte, valueLength)
	value1[valueLength-1] = 4
	value2 := make([]byte, valueLength)
	value2[valueLength-1] = 8
	value3 := make([]byte, valueLength)
	value3[valueLength-1] = 16
	values := make([][]byte, 4)
	values[0] = CopySlice(value0)
	values[1] = CopySlice(value1)
	values[2] = CopySlice(value2)
	values[3] = CopySlice(value3)

	s := kvs{keys: keys, values: values}
	if s.Len() != 4 {
		t.Fatal("Invalid length")
	}
	lessThan := s.Less(0, 1)
	if lessThan {
		t.Fatal("key0 shoud not be less than key1")
	}
	s.Swap(2, 3)
	if !bytes.Equal(s.keys[2], key3) {
		t.Fatal("Did not correctly swap key 2")
	}
	if !bytes.Equal(s.values[2], value3) {
		t.Fatal("Did not correctly swap value 2")
	}
	if !bytes.Equal(s.keys[3], key2) {
		t.Fatal("Did not correctly swap key 3")
	}
	if !bytes.Equal(s.values[3], value2) {
		t.Fatal("Did not correctly swap value 3")
	}

	valuesBad := values[:3]
	_, _, err := SortKVs(keys, valuesBad)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	keysSorted, valuesSorted, err := SortKVs(s.keys, s.values)
	if err != nil {
		t.Fatal(err)
	}
	// Correct sorting: 1, 3, 2, 0
	if !bytes.Equal(keysSorted[0], key1) {
		t.Fatal("Invalid sort: should be key1")
	}
	if !bytes.Equal(valuesSorted[0], value1) {
		t.Fatal("Invalid sort: should be value1")
	}
	if !bytes.Equal(keysSorted[1], key3) {
		t.Fatal("Invalid sort: should be key3")
	}
	if !bytes.Equal(valuesSorted[1], value3) {
		t.Fatal("Invalid sort: should be value3")
	}
	if !bytes.Equal(keysSorted[2], key2) {
		t.Fatal("Invalid sort: should be key2")
	}
	if !bytes.Equal(valuesSorted[2], value2) {
		t.Fatal("Invalid sort: should be value2")
	}
	if !bytes.Equal(keysSorted[3], key0) {
		t.Fatal("Invalid sort: should be key0")
	}
	if !bytes.Equal(valuesSorted[3], value0) {
		t.Fatal("Invalid sort: should be value0")
	}
}
