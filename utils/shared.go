package utils

import (
	"crypto/rand"
	"errors"
	"math/big"
	"os/user"
	"path/filepath"

	"github.com/dgraph-io/badger/v2"
	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/errorz"
	"github.com/alicenet/alicenet/logging"
)

// ForceSliceToLength will return a byte slice of size length.
// It will left pad a byte slice to the specified number of zeros if the
// slice is not long enough. If the slice is too long, it will return the
// right-most bytes of the slice.
func ForceSliceToLength(inSlice []byte, length int) []byte {
	if len(inSlice) > length {
		return CopySlice(inSlice[len(inSlice)-length:])
	}
	outSlice := make([]byte, length-len(inSlice))
	outSlice = append(outSlice, CopySlice(inSlice)...)
	return outSlice
}

// CopySlice returns a copy of a passed byte slice.
func CopySlice(v []byte) []byte {
	out := make([]byte, len(v))
	copy(out, v)
	return out
}

// Epoch returns the epoch for the corresponding height.
func Epoch(height uint32) uint32 {
	if height <= constants.EpochLength {
		return 1
	}
	if height%constants.EpochLength == 0 {
		return height / constants.EpochLength
	}
	return (height / constants.EpochLength) + 1
}

// ValidateHash checks whether or not hsh has the correct length.
func ValidateHash(hsh []byte) error {
	if len(hsh) != constants.HashLen {
		return errorz.ErrInvalid{}.New("the length of the hash is incorrect")
	}
	return nil
}

// RandomBytes will return a byte slice of num random bytes using crypto rand.
func RandomBytes(num int) ([]byte, error) {
	b := make([]byte, num)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// OpenBadger opens a badgerdb database and closes the db when closeChan
// returns a struct{}{}.
func OpenBadger(closeChan <-chan struct{}, directoryName string, inMemory bool) (*badger.DB, error) {
	logger := logging.GetLogger(constants.LoggerBadger)

	if len(directoryName) >= 2 {
		if directoryName[0:2] == "~/" {
			usr, err := user.Current()
			if err != nil {
				return nil, err
			}
			directoryName = filepath.Join(usr.HomeDir, directoryName[1:])
			logger.Infof("Directory:%v", directoryName)
		}
	}

	logger.Infof("Opening badger DB... In-Memory:%v Directory:%v", inMemory, directoryName)
	opts := badger.DefaultOptions(directoryName).WithInMemory(inMemory).WithSyncWrites(true)
	opts.Logger = logger

	thisDB, err := badger.Open(opts)
	if err != nil {
		logger.Errorf("Could not open database: %v", err)
		return nil, err
	}
	go func() {
		<-closeChan
		thisDB.Close()
	}()
	// if err := thisDB.Flatten(4); err != nil {
	// 	return nil, err
	// }
	// if !inMemory {
	// 	for {
	// 		if err := thisDB.RunValueLogGC(constants.BadgerDiscardRatio); err != nil {
	// 			if err == badger.ErrNoRewrite {
	// 				break
	// 			}
	// 			return nil, err
	// 		}
	// 	}
	// }
	return thisDB, nil
}

// DebugTrace allows a traceback to be generated that includes a file name,
// a line number, the error message, and an optional string. Calling this
// function using a logger that is set to anthything other than trace or debug
// level is a no-op. This filtering helps to minimize overhead during normal
// use but still allows error tracebacks to be created easily. The returned
// file and line number will point to where  this function was called.
// Although more than one string may be passed, only the first string will
// be displayed. The varadic property was only used to shorten calling syntax.
func DebugTrace(logger *logrus.Logger, err error, s ...string) {
	// TODO: make more generic, e.g. DebugTrace(l,err); DebugTrace(l,str); DebugTrace(l,pattern,v,v)
	if logger.GetLevel() > logrus.DebugLevel {
		return
	}

	trace := errorz.MakeTrace(1)

	if err != nil {
		if len(s) > 0 {
			logger.WithField("l", trace).Debugf("%v ::: %v", err.Error(), s[0])
			return
		}
		logger.WithField("l", trace).Debugf("%v", err.Error())
		return
	}
	if len(s) > 0 {
		logger.WithField("l", trace).Debugf("%v", s[0])
		return
	}
	logger.WithField("l", trace).Debug("")
}

// CloneBigInt makes a deep copy of a big.Int.
func CloneBigInt(original *big.Int) *big.Int {
	return new(big.Int).Set(original)
}

// CloneSliceBigInt2 makes a deep copy of a slice of array [2] of big.Int's.
func CloneSliceBigInt2(original [][2]*big.Int) [][2]*big.Int {
	clone := make([][2]*big.Int, len(original))
	for idx := 0; idx < len(original); idx++ {
		clone[idx] = CloneBigInt2(original[idx])
	}
	return clone
}

// CloneBigInt2 makes a deep copy of an array [2] of big.Int's.
func CloneBigInt2(original [2]*big.Int) [2]*big.Int {
	return [2]*big.Int{new(big.Int).Set(original[0]), new(big.Int).Set(original[1])}
}

// CloneBigInt4 makes a deep copy of a array [4] of big.Int's.
func CloneBigInt4(original [4]*big.Int) [4]*big.Int {
	return [4]*big.Int{
		new(big.Int).Set(original[0]),
		new(big.Int).Set(original[1]),
		new(big.Int).Set(original[2]),
		new(big.Int).Set(original[3]),
	}
}

// CloneBigIntSlice makes a deep copy of a slice of big.Int's.
func CloneBigIntSlice(original []*big.Int) []*big.Int {
	clone := make([]*big.Int, len(original))
	for idx := 0; idx < len(original); idx++ {
		clone[idx] = CloneBigInt(original[idx])
	}
	return clone
}

// Computes and returns the max between two uint64.
func Max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// Computes and returns the min between two uint64.
func Min(a, b uint64) uint64 {
	if a > b {
		return b
	}
	return a
}

// StringToBytes32 is useful for convert a Go string into a bytes32 useful calling Solidity.
func StringToBytes32(str string) (b [32]byte) {
	copy(b[:], []byte(str)[0:32])
	return
}

// HandleBadgerErrors handles the badger errors. This function will suppress the
// errors that we expect to happens and are not harmful to the system.
func HandleBadgerErrors(err error) error {
	if !errors.Is(err, badger.ErrNoRewrite) && !errors.Is(err, badger.ErrGCInMemoryMode) {
		return err
	}
	return nil
}
