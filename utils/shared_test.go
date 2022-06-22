package utils

import (
	"bytes"
	"context"
	"errors"
	"github.com/alicenet/alicenet/logging"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"time"

	"github.com/alicenet/alicenet/constants"

	"testing"
)

func TestUtils(t *testing.T) {
	for i := uint32(1); i <= (constants.EpochLength * 4); i++ {
		epoch := Epoch(i)
		switch {
		case i < constants.EpochLength:
			if epoch != 1 {
				t.Fatal(i, epoch)
			}
		case i == constants.EpochLength:
			if epoch != 1 {
				t.Fatal(i, epoch)
			}
		case i == constants.EpochLength+1:
			if epoch != 2 {
				t.Fatal(i, epoch)
			}
		case i < constants.EpochLength*2:
			if epoch != 2 {
				t.Fatal(i, epoch)
			}
		case i == constants.EpochLength*2:
			if epoch != 2 {
				t.Fatal(i, epoch)
			}
		case i == constants.EpochLength*2+1:
			if epoch != 3 {
				t.Fatal(i, epoch)
			}
		case i < constants.EpochLength*3:
			if epoch != 3 {
				t.Fatal(i, epoch)
			}
		case i == constants.EpochLength*3:
			if epoch != 3 {
				t.Fatal(i, epoch)
			}
		case i == constants.EpochLength*3+1:
			if epoch != 4 {
				t.Fatal(i, epoch)
			}
		case i < constants.EpochLength*4:
			if epoch != 4 {
				t.Fatal(i, epoch)
			}
		case i == constants.EpochLength*4:
			if epoch != 4 {
				t.Fatal(i, epoch)
			}
		}
	}
}

func TestCopySlice(t *testing.T) {
	zeroBytes := make([]byte, 10)
	zeroBytesCopy := CopySlice(zeroBytes)
	if !bytes.Equal(zeroBytes, zeroBytesCopy) {
		t.Fatal("Byte slices do not agree!")
	}
}

func TestPadZeros(t *testing.T) {
	ones := make([]byte, 8)
	for i := 0; i < len(ones); i++ {
		ones[i] = 255
	}
	res := ForceSliceToLength(ones, 16)
	resTrue := make([]byte, 16)

	resTrue[8] = 255
	resTrue[9] = 255
	resTrue[10] = 255
	resTrue[11] = 255
	resTrue[12] = 255
	resTrue[13] = 255
	resTrue[14] = 255
	resTrue[15] = 255

	if !bytes.Equal(res, resTrue) {
		t.Fatal("PadZeros failed!")
	}
}

func TestTruncate(t *testing.T) {
	badBytes := make([]byte, 34)
	badBytes[len(badBytes)-1] = uint8(255)
	padSmall := 32
	defer func() {
		if err := recover(); err != nil {
			t.Fatal("Should not cause panic!")
		}
	}()
	after := ForceSliceToLength(badBytes, padSmall)
	if after[len(after)-1] != 255 {
		t.Fatal("Incorrect value after truncation")
	}
	if len(after) != padSmall {
		t.Fatal("Incorrect length after truncation")
	}
}

func TestValidateHash(t *testing.T) {
	hash := make([]byte, constants.HashLen)
	err := ValidateHash(hash)
	if err != nil {
		t.Fatal(err)
	}
	hashBad := make([]byte, constants.HashLen+1)
	err = ValidateHash(hashBad)
	if err == nil {
		t.Fatal("Should have raised error: invalid hash length")
	}
}

func TestRandomBytes(t *testing.T) {
	n := 32
	bt, err := RandomBytes(n)
	assert.Nil(t, err)
	assert.Equal(t, n, len(bt))
}

func TestOpenBadger(t *testing.T) {
	dbPath := "~/test/"

	ctxTimeout := 10 * time.Millisecond
	ctx := context.Background()
	nodeCtx, cf := context.WithTimeout(ctx, ctxTimeout)
	defer cf()

	db, err := OpenBadger(
		nodeCtx.Done(),
		dbPath,
		false,
	)

	assert.Nil(t, err)
	time.Sleep(ctxTimeout)
	assert.True(t, db.IsClosed())
}

func TestDebugTrace(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code shouldn't panic")
		}
	}()
	logger := logging.GetLogger("test")
	logger.SetLevel(logrus.DebugLevel)

	DebugTrace(logger, errors.New("test error"))
	DebugTrace(logger, errors.New("test error"), "test string")
	DebugTrace(logger, nil, "test string")
	DebugTrace(logger, nil)
}
