package dynamics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateGet(t *testing.T) {
	update := &Update{}
	if len(update.Value()) != 0 {
		t.Fatal("Should have raised error (1)")
	}
	if update.Epoch() != 0 {
		t.Fatal("Should have raised error (2)")
	}
}

func TestUpdateNewUpdate(t *testing.T) {
	value := []byte("123456789")
	epoch := uint32(1)
	update := NewUpdate(value, epoch)
	assert.Equal(t, update.Value(), value)
	assert.Equal(t, update.Epoch(), epoch)
}
