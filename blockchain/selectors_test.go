package blockchain_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/alicenet/alicenet/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestSelector_Selector(t *testing.T) {
	sm := blockchain.NewSelectorMap()

	selector := sm.Selector("fdsfds")

	assert.NotEqual(t, []byte{0, 0, 0, 0}, selector)
}

func TestSelector_Signature(t *testing.T) {
	sm := blockchain.NewSelectorMap()

	testSig := "fdsfds"

	selector := sm.Selector(testSig)

	signature := sm.Signature(selector)

	assert.Equal(t, testSig, signature)
}

func TestSelector_Concurrency(t *testing.T) {

	sm := blockchain.NewSelectorMap()
	iter := 10000
	n := 10
	wg := &sync.WaitGroup{}

	match := make([]bool, n)
	fn := func(id int, wg *sync.WaitGroup) {
		for idx := 0; idx < iter; idx++ {
			testSig := fmt.Sprintf("fn%v()", idx)
			selector := sm.Selector(testSig)
			signature := sm.Signature(selector)

			if testSig != signature {
				match[id] = false
			}
		}
		wg.Done()
	}

	for idx := 0; idx < n; idx++ {
		match[idx] = true
		wg.Add(1)
		go fn(idx, wg)
	}

	wg.Wait()

	for idx := 0; idx < n; idx++ {
		assert.True(t, match[idx])
	}
}
