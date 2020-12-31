package monitor_test

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/MadBase/MadNet/blockchain/monitor"
	"github.com/stretchr/testify/assert"
)

func TestBidirectionalMarshal(t *testing.T) {

	// Build up a pseudo-realistic State instance
	ms := &monitor.State{}
	ms.Version = 0
	ms.HighestBlockProcessed = 614
	ms.HighestBlockFinalized = 911
	ms.HighestEpochProcessed = 5
	ms.HighestEpochSeen = 10
	ms.LatestDepositProcessed = 1
	ms.LatestDepositSeen = 5
	ms.ValidatorSets = make(map[uint32]monitor.ValidatorSet)
	ms.Validators = make(map[uint32][]monitor.Validator)
	ms.Validators[614] = []monitor.Validator{{Index: 7}}

	// Encode the test instance
	buf := &bytes.Buffer{}
	enc := gob.NewEncoder(buf)
	err := enc.Encode(ms)
	assert.Nilf(t, err, "Should be no errors marshalling data")

	// Decode the bytes
	ms2 := &monitor.State{}
	dec := gob.NewDecoder(buf)
	err = dec.Decode(ms2)
	assert.Nilf(t, err, "Should be no errors unmarshalling data")

	// Make sure the new struct looks like the old struct
	assert.Equal(t, uint64(614), ms2.HighestBlockProcessed)
	assert.Equal(t, uint64(911), ms2.HighestBlockFinalized)
	assert.Equal(t, uint32(5), ms2.HighestEpochProcessed)
	assert.Equal(t, uint32(10), ms2.HighestEpochSeen)
	assert.Equal(t, uint32(5), ms2.HighestEpochProcessed)
	assert.Equal(t, uint32(1), ms2.LatestDepositProcessed)
	assert.Equal(t, uint32(5), ms2.LatestDepositSeen)
	assert.Equal(t, uint8(7), ms2.Validators[614][0].Index)
}
