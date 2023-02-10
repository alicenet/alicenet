package marshaller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Common interface.
type S interface {
	Do() int
}

// Structs.
type S0 struct {
	Val int
}

type S1 struct {
	Val int
}

type S2 struct {
	Val int
}

// Stupid receivers.
func (s *S0) Do() int {
	return s.Val * 2
}

func (s *S1) Do() int {
	return s.Val * 3
}

func TestRoundTrip(t *testing.T) {
	tr := &TypeRegistry{}

	s0 := &S0{Val: 3}
	s1 := &S1{Val: 4}

	tr.RegisterInstanceType(&S0{})
	tr.RegisterInstanceType(&S1{})

	// Wrap both
	ws0, err := tr.WrapInstance(s0)
	assert.Nil(t, err)
	assert.NotNil(t, ws0)

	ws1, err := tr.WrapInstance(s1)
	assert.Nil(t, err)
	assert.NotNil(t, ws1)

	// Unwrap both
	uw0, err := tr.UnwrapInstance(ws0)
	assert.Nil(t, err)

	uw1, err := tr.UnwrapInstance(ws1)
	assert.Nil(t, err)

	uws0, ok := uw0.(S)
	if !ok {
		t.Errorf("Unable to cast marshalled value")
	}
	uws1, ok := uw1.(S)
	if !ok {
		t.Errorf("Unable to cast marshalled value")
	}

	// Make sure everything worked
	assert.Equal(t, 6, uws0.Do())
	assert.Equal(t, 12, uws1.Do())
}

func TestNegativeType(t *testing.T) {
	defer func() {
		// If we didn't get here by recovering from a panic() we failed
		if reason := recover(); reason == nil {
			t.Log("No panic in sight")
			t.Fatal("Should have panicked")
		} else {
			t.Logf("Good panic because: %v", reason)
		}
	}()

	tr := &TypeRegistry{}

	s0 := &S0{Val: 3}
	s1 := &S1{Val: 4}
	s2 := &S2{Val: 5}

	tr.RegisterInstanceType(&S0{})
	tr.RegisterInstanceType(&S1{})

	var err error

	_, err = tr.WrapInstance(s0)
	assert.Nil(t, err)

	_, err = tr.WrapInstance(s1)
	assert.Nil(t, err)

	_, err = tr.WrapInstance(s2)
	assert.Equal(t, ErrUnknownType, err)
}

func TestNegativeName(t *testing.T) {
	tr := &TypeRegistry{}

	s0 := &S0{Val: 3}

	tr.RegisterInstanceType(&S0{})

	wi, err := tr.WrapInstance(s0)
	assert.Nil(t, err)

	wi.NameType = "bob"
	_, err = tr.UnwrapInstance(wi)
	assert.Equal(t, ErrUnknownName, err)
}

func TestNegativeRaw(t *testing.T) {
	tr := &TypeRegistry{}

	s0 := &S0{Val: 3}

	tr.RegisterInstanceType(&S0{})

	wi, err := tr.WrapInstance(s0)
	assert.Nil(t, err)

	wi.RawInstance = []byte{'f', 'o', 'o'}
	_, err = tr.UnwrapInstance(wi)
	assert.NotNil(t, err)
}
