package blockchain_test

import (
	"testing"
	"time"

	"github.com/golang-collections/go-datastructures/queue"
	"github.com/stretchr/testify/assert"
)

func TestEnqueue(t *testing.T) {
	q := queue.New(3)
	assert.Nil(t, q.Put(1))
	assert.Nil(t, q.Put(2))
	assert.Nil(t, q.Put(3))
	assert.Nil(t, q.Put(4))

	assert.Equal(t, int64(4), q.Len())

	go func() {
		time.Sleep(5 * time.Second)
		err := q.Put(5)
		assert.Nil(t, err)
	}()

	for i := 0; i < 5; i++ {
		val, err := q.Get(1)
		assert.Nil(t, err)
		assert.Equal(t, i+1, val[0].(int))
	}
}
