package peering

import (
	"crypto/rand"
	"errors"
	"math/big"
)

func makePid() uint64 {
	bigZero := big.NewInt(0)
	maxSizeBig := big.NewInt(1)
	maxSizeBig.Lsh(maxSizeBig, 64)
	for {
		pidBig, err := rand.Int(rand.Reader, maxSizeBig)
		// Note: Previous code forced value to be strictly positive.
		// Strictly positive means that all nontrivial map elements will
		// be nonzero.
		if err != nil || pidBig.Cmp(bigZero) == 0 {
			continue
		}
		pid := pidBig.Uint64()
		return pid
	}
}

func randomElement(maxSize int) (int, error) {
	if maxSize <= 0 {
		return 0, errors.New("invalid randomElements arg: maxSize <= 0")
	}
	if maxSize == 1 {
		return 0, nil
	}
	maxSizeBig := big.NewInt(int64(maxSize))
	for {
		idxBig, err := rand.Int(rand.Reader, maxSizeBig)
		if err != nil {
			continue
		}
		idx := int(idxBig.Int64())
		return idx, nil
	}
}
