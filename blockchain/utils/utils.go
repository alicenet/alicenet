package utils

import (
	"context"
	"math/big"
	"time"
)

// SleepWithContext sleeps for a specified duration, unless the provided context completes earlier
func SleepWithContext(ctx context.Context, delay time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(delay):
		return nil
	}
}

// SlowReturn combines Sleep with returning specified bool value
func SlowReturn(ctx context.Context, delay time.Duration, value bool) (bool, error) {
	err := SleepWithContext(ctx, delay)
	if err != nil {
		return false, err
	}
	return value, nil
}

// CloneBigInt makes a deep copy of a big.Int
func CloneBigInt(original *big.Int) *big.Int {
	return new(big.Int).Set(original)
}

// CloneSliceBigInt2 makes a deep copy of a slice of array [2] of big.Int's
func CloneSliceBigInt2(original [][2]*big.Int) [][2]*big.Int {
	clone := make([][2]*big.Int, len(original))
	for idx := 0; idx < len(original); idx++ {
		clone[idx] = CloneBigInt2(original[idx])
	}
	return clone
}

// CloneBigInt2 makes a deep copy of an array [2] of big.Int's
func CloneBigInt2(original [2]*big.Int) [2]*big.Int {
	return [2]*big.Int{new(big.Int).Set(original[0]), new(big.Int).Set(original[1])}
}

// CloneBigInt4 makes a deep copy of a array [4] of big.Int's
func CloneBigInt4(original [4]*big.Int) [4]*big.Int {
	return [4]*big.Int{
		new(big.Int).Set(original[0]),
		new(big.Int).Set(original[1]),
		new(big.Int).Set(original[2]),
		new(big.Int).Set(original[3])}
}

// CloneBigIntSlice makes a deep copy of a slice of big.Int's
func CloneBigIntSlice(original []*big.Int) []*big.Int {
	clone := make([]*big.Int, len(original))
	for idx := 0; idx < len(original); idx++ {
		clone[idx] = CloneBigInt(original[idx])
	}
	return clone
}

// Computes and returns the max between two uint64
func Max(a uint64, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// Computes and returns the min between two uint64
func Min(a uint64, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}
