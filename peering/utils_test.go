package peering

import (
	"testing"
)

func TestMakePid(t *testing.T) {
	_ = makePid()
}

func TestRandomElement(t *testing.T) {
	maxSize := 0
	_, err := randomElement(maxSize)
	if err == nil {
		t.Fatal("Should have raised error")
	}
	maxSize = 1
	val, err := randomElement(maxSize)
	if err != nil {
		t.Fatal(err)
	}
	if val != 0 {
		t.Fatal("val should be zero")
	}

	num := 1000
	maxSize = 10
	for i := 0; i < num; i++ {
		val, err := randomElement(maxSize)
		if err != nil {
			t.Fatal(err)
		}
		if val < 0 || val >= maxSize {
			t.Fatal("Invalid randomElement call")
		}
	}
}
