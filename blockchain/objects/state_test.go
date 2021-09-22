package objects_test

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/stretchr/testify/assert"
)

func TestBidirectionalJson(t *testing.T) {

	// Build up a pseudo-realistic State instance
	ms := createState()

	// Encode the test instance
	raw, err := json.Marshal(ms)
	assert.Nilf(t, err, "Should be no errors marshalling data")

	t.Logf("raw:%v", string(raw))

	// Decode the bytes
	ms2 := &objects.MonitorState{}
	err = json.Unmarshal(raw, ms2)
	assert.Nilf(t, err, "Should be no errors unmarshalling data")

	// Good?
	assertStateMatch(t, ms2)
}

func createState() *objects.MonitorState {

	ms := &objects.MonitorState{
		Version:                0,
		HighestBlockProcessed:  614,
		HighestBlockFinalized:  911,
		HighestEpochProcessed:  5,
		HighestEpochSeen:       10,
		LatestDepositProcessed: 1,
		LatestDepositSeen:      5,
		ValidatorSets:          map[uint32]objects.ValidatorSet{},
		Validators:             map[uint32][]objects.Validator{614: {{Index: 7}}},
	}

	return ms
}

func assertStateMatch(t *testing.T, ms *objects.MonitorState) {
	// Make sure the new struct looks like the old struct
	assert.Equal(t, uint64(614), ms.HighestBlockProcessed)
	assert.Equal(t, uint64(911), ms.HighestBlockFinalized)
	assert.Equal(t, uint32(5), ms.HighestEpochProcessed)
	assert.Equal(t, uint32(10), ms.HighestEpochSeen)
	assert.Equal(t, uint32(5), ms.HighestEpochProcessed)
	assert.Equal(t, uint32(1), ms.LatestDepositProcessed)
	assert.Equal(t, uint32(5), ms.LatestDepositSeen)
	assert.Equal(t, uint8(7), ms.Validators[614][0].Index)
}

func TestFubar(t *testing.T) {

	var m map[int]string
	// var o sync.Once

	fn := func(n int) string {
		return m[n]
	}

	// m := map[int]string{0: "bob", 1: "ross", 2: "foo", 3: "bar", 4: "Ah!"}

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 100; j++ {
				fmt.Println(fn(j % 5))
				time.Sleep(100 * time.Millisecond)
			}
			wg.Done()
		}()
	}

	wg.Wait()
}
