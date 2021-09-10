package errorz

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrInvalidWrapping(t *testing.T) {
	err := NewErrConsensus("something low level went wrong", true)
	err2 := NewErrInvalid("something high level went wrong").Wrap(err)

	fmt.Println(err2.Error())
	if err2.Error() != "the object is invalid: something high level went wrong:\nsomething low level went wrong" {
		t.Fatal("error message of wrapper error should be properly concatenated with wrapped error")
	}

	var wrapped *ErrConsensus
	if !errors.As(err2, &wrapped) {
		t.Fatal("err2 should unwrap to err using errors.As")
	}

	if !wrapped.IsLocal() {
		t.Fatal("wrapped attributes should be readable")
	}

	if wrapped.Error() != "something low level went wrong" {
		t.Fatal("wrapped Error() should return proper error string")
	}

}

func TestErrInvalidUnwrapped(t *testing.T) {
	err := ErrInvalid{}.New("something high level went wrong")

	if err.Error() != "the object is invalid: something high level went wrong" {
		t.Fatal("ErrInvalid needs proper Error() implementation when unwrapped")
	}

}
