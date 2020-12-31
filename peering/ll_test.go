package peering

import (
	"testing"
)

func TestLL(t *testing.T) {
	ll := newLinkedList(4)
	evictions := ll.Push("one", "two", "three", "four")
	if len(evictions) > 0 {
		t.Fatal("fail")
	}
	if !ll.Contains("one") {
		t.Fatal("fail")
	}
	if !ll.Contains("two") {
		t.Fatal("fail")
	}
	if !ll.Contains("three") {
		t.Fatal("fail")
	}
	if !ll.Contains("four") {
		t.Fatal("fail")
	}
	evictions = ll.Push("one", "two", "three", "five")
	if !ll.Contains("one") {
		t.Fatal("fail")
	}
	if !ll.Contains("two") {
		t.Fatal("fail")
	}
	if !ll.Contains("three") {
		t.Fatal("fail")
	}
	if ll.Contains("four") {
		t.Fatal("fail")
	}
	if !ll.Contains("five") {
		t.Fatal("fail")
	}
	if len(evictions) != 1 {
		t.Fatal("fail", len(evictions))
	}
	if evictions[0] != "four" {
		t.Fatal("fail")
	}
}
