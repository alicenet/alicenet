package blockchain_test

import (
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

const SLEEP_DURATION = 500 * time.Millisecond

func TestSleepWithContextComplete(t *testing.T) {
	ctx, _ := context.WithCancel(context.Background())

	completed := false

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := blockchain.SleepWithContext(ctx, SLEEP_DURATION)
		if err == nil {
			completed = true
		}
		wg.Done()
	}()

	wg.Wait()

	assert.True(t, completed)
}

func TestSleepWithContextInterupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	completed := false

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		err := blockchain.SleepWithContext(ctx, SLEEP_DURATION)
		if err == nil {
			completed = true
		} else {
			t.Logf("sleep interupted: %v", err)
		}
		wg.Done()
	}()

	cancel()

	wg.Wait()

	assert.False(t, completed)
}
