package errorz

import (
	"errors"
	"testing"
)

func TestErrInvalidWrapping(t *testing.T) {
	err := NewErrConsensus("something low level went wrong", true)
	err2 := WrapErrInvalid(err, "something high level went wrong")

	var wrapped *ErrConsensus
	if !errors.As(err2, &wrapped) {
		t.Fatal("err2 should contain err")
	}

	if !wrapped.isLocal {
		t.Fatal("isLocal should be true")
	}
}
