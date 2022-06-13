package blockchain_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/stretchr/testify/assert"
)

const SLEEP_DURATION = 500 * time.Millisecond

func TestCancellation_SleepWithContextComplete(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	completed := false

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := blockchain.SleepWithContext(ctx, SLEEP_DURATION)
		if err == nil {
			completed = true
		}
	}()
	wg.Wait()

	assert.True(t, completed)
}

func TestCancellation_SleepWithContextInterrupted(t *testing.T) {
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

func TestCancellation_SlowReturn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	type args struct {
		ctx   context.Context
		delay time.Duration
		value bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test slow Return success",
			args: args{
				ctx:   ctx,
				delay: 5000 * time.Millisecond,
				value: true,
			},
			want: true,
		},
		{
			name: "Test slow Return quicker delay",
			args: args{
				ctx:   ctx,
				delay: 500 * time.Millisecond,
				value: false,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()
			slowReturn, _ := blockchain.SlowReturn(tt.args.ctx, tt.args.delay, tt.args.value)
			assert.Equalf(t, tt.want, slowReturn, "SlowReturn(%v, %v, %v)", tt.args.ctx, tt.args.delay, tt.args.value)
			elapsed := time.Since(start)
			assert.GreaterOrEqual(t, elapsed, tt.args.delay, "Delay time has not been respected")
		})
	}
}
