package lstate

import (
	"testing"
	"time"
)

func TestDLCAddCancelOne(t *testing.T) {
	shortTO := 2 * time.Second
	cache := &dLCache{}
	err := cache.init(shortTO)
	if err != nil {
		t.Fatal(err)
	}
	txHsh := []byte("a")
	err = cache.add(txHsh)
	if err != nil {
		t.Fatal(err)
	}
	present := cache.containsTxHsh(txHsh)
	if !present {
		t.Fatal("Not present!")
	}
	present = cache.cancelOne(txHsh)
	if !present {
		t.Fatal("Present when should not be!")
	}
	present = cache.containsTxHsh(txHsh)
	if present {
		t.Fatal("Should not be present!")
	}
}

func TestExtend(t *testing.T) {
	shortTO := 2 * time.Second
	cache := &dLCache{}
	err := cache.init(shortTO)
	if err != nil {
		t.Fatal(err)
	}
	txHsh := []byte("a")
	err = cache.add(txHsh)
	if err != nil {
		t.Fatal(err)
	}
	present := cache.containsTxHsh(txHsh)
	if !present {
		t.Fatal("Not present!")
	}
	time.Sleep(1 * time.Second)
	err = cache.add(txHsh)
	if err != nil {
		t.Fatal(err)
	}
	deadline := cache.get(txHsh)
	if !time.Now().Add(-1 * time.Second).Before(deadline) {
		t.Fatal("bad deadline")
	}
}

func TestDLCCancelAll(t *testing.T) {
	shortTO := 2 * time.Second
	cache := &dLCache{}
	err := cache.init(shortTO)
	if err != nil {
		t.Fatal(err)
	}
	txHsh := []byte("a")
	err = cache.add(txHsh)
	if err != nil {
		t.Fatal(err)
	}
	present := cache.containsTxHsh(txHsh)
	if !present {
		t.Fatal("a not present!")
	}
	txHsh2 := []byte("b")
	err = cache.add(txHsh2)
	if err != nil {
		t.Fatal(err)
	}
	present = cache.containsTxHsh(txHsh2)
	if !present {
		t.Fatal("b not present!")
	}
	cache.cancelAll()
	present = cache.containsTxHsh(txHsh)
	if present {
		t.Fatal("a present when should not be!")
	}
	present = cache.containsTxHsh(txHsh2)
	if present {
		t.Fatal("b present when should not be!")
	}
}

func TestDLCExpired(t *testing.T) {
	shortTO := 2 * time.Second
	cache := &dLCache{}
	err := cache.init(shortTO)
	if err != nil {
		t.Fatal(err)
	}
	txHsh := []byte("a")
	err = cache.add(txHsh)
	if err != nil {
		t.Fatal(err)
	}
	present := cache.containsTxHsh(txHsh)
	if !present {
		t.Fatal("a not present!")
	}
	exp := cache.expired(txHsh)
	if exp {
		t.Fatal("Should not be expired!")
	}
	time.Sleep(1*time.Millisecond + shortTO)
	exp = cache.expired(txHsh)
	if !exp {
		t.Fatal("Should have expired!")
	}
}
