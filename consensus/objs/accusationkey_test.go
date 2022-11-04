package objs

import (
	"testing"

	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/errorz"
	"github.com/stretchr/testify/assert"
)

func TestAccusationKeyMarshalBinary(t *testing.T) {
	// create a new AccusationKey
	var id [32]byte
	copy(id[:], crypto.Hasher([]byte("testId")))
	acc := &AccusationKey{
		Prefix: []byte("testPrefix"),
		ID:     id,
	}

	// marshal it
	bin, err := acc.MarshalBinary()
	assert.Nil(t, err)
	assert.NotEmpty(t, bin)

	// unmarshal it to a new AccusationKey variable
	acc2 := &AccusationKey{}
	err = acc2.UnmarshalBinary(bin)
	assert.Nil(t, err)

	// make sure the AccusationKey variables are the same
	assert.Equal(t, acc, acc2)
	assert.Equal(t, acc.Prefix, acc2.Prefix)
	assert.Equal(t, acc.ID, acc2.ID)
}

func TestAccusationKeyMakeIterKey(t *testing.T) {
	// create a new AccusationKey
	var id [32]byte
	copy(id[:], crypto.Hasher([]byte("testId")))
	acc := &AccusationKey{
		Prefix: []byte("testPrefix"),
		ID:     id,
	}

	// MakeIterKey it
	bin, err := acc.MakeIterKey()
	assert.Nil(t, err)
	assert.NotEmpty(t, bin)
}

func TestAccusationKeyMarshalNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should have panicked")
		}
	}()

	// create a nil AccusationKey pointer
	var acc *AccusationKey

	// marshal it
	acc.MarshalBinary() // panics
}

func TestAccusationKeyUnmarshalNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should have panicked")
		}
	}()

	// create a new AccusationKey
	var id [32]byte
	copy(id[:], crypto.Hasher([]byte("testId")))
	acc := &AccusationKey{
		Prefix: []byte("testPrefix"),
		ID:     id,
	}

	// marshal it
	bin, err := acc.MarshalBinary()
	assert.Nil(t, err)
	assert.NotEmpty(t, bin)

	// create a nil AccusationKey pointer
	var acc2 *AccusationKey

	acc2.UnmarshalBinary(bin) // panics
}

func TestAccusationKeyMakeIterKeyNil(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("should have panicked")
		}
	}()

	// create a nil AccusationKey pointer
	var acc *AccusationKey

	// MakeIterKey it
	acc.MakeIterKey() // panics
}

func TestAccusationKeyUnmarshalWrongBinary(t *testing.T) {
	// create a new AccusationKey
	var id [32]byte
	copy(id[:], crypto.Hasher([]byte("testId")))
	acc := &AccusationKey{
		Prefix: []byte("testPrefix"),
		ID:     id,
	}

	// marshal it
	bin, err := acc.MarshalBinary()
	assert.Nil(t, err)
	assert.NotEmpty(t, bin)

	// add more data to bin
	bin = append(bin, []byte("|")...)
	bin = append(bin, []byte("123")...)

	// unmarshal it to a new AccusationKey variable
	acc2 := &AccusationKey{}
	err = acc2.UnmarshalBinary(bin)
	assert.NotNil(t, err)
	assert.Equal(t, errorz.ErrCorrupt, err)
}
