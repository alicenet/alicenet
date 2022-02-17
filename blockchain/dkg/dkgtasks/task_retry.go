package dkgtasks

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/utils"
	"github.com/sirupsen/logrus"
)

// GeneralTaskShouldRetry is the general logic used to determine if a task should try again
// -- Process is
func GeneralTaskShouldRetry(ctx context.Context, logger *logrus.Entry,
	eth interfaces.Ethereum, expectedFirstBlock uint64, expectedLastBlock uint64) bool {

	result := internalGeneralTaskShouldRetry(ctx, logger, eth, expectedFirstBlock, expectedLastBlock)

	logger.Infof("GeneralTaskRetry(expectedFirstBlock: %v expectedLastBlock:%v): %v",
		expectedFirstBlock,
		expectedLastBlock,
		result)

	return result
}

// FormatPublicKey formats the public key suitably for logging
func FormatPublicKey(publicKey [2]*big.Int) string {
	pk0BytesRaw := publicKey[0].Bytes()
	pk1BytesRaw := publicKey[1].Bytes()
	pk0Bytes := utils.ForceSliceToLength(pk0BytesRaw, 32)
	pk1Bytes := utils.ForceSliceToLength(pk1BytesRaw, 32)
	pk0Hex := utils.EncodeHexString(pk0Bytes)
	pk1Hex := utils.EncodeHexString(pk1Bytes)
	pk0 := pk0Hex[0:3]
	pk1 := pk1Hex[len(pk1Hex)-3:]
	return fmt.Sprintf("0x%v...%v", pk0, pk1)
}

// FormatBigIntSlice formats a slice of *big.Int's suitably for logging
func FormatBigIntSlice(slice []*big.Int) string {
	var b bytes.Buffer
	for _, i := range slice {
		b.WriteString(i.Text(16))
	}

	str := b.String()

	if len(str) < 16 {
		return fmt.Sprintf("0x%v", str)
	}

	return fmt.Sprintf("0x%v...%v", str[0:3], str[len(str)-3:])
}

func internalGeneralTaskShouldRetry(ctx context.Context, logger *logrus.Entry,
	eth interfaces.Ethereum, expectedFirstBlock uint64, expectedLastBlock uint64) bool {

	// Make sure we're in the right block range to continue
	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		// This probably means an endpoint issue, so we have to try again
		logger.Warnf("could not check current height of chain: %v", err)
		return true
	}

	if currentBlock >= expectedLastBlock {
		return false
	}

	// TODO Any other general cases where we know retry won't work?

	// We won't loop forever because eventually currentBlock > expectedLastBlock
	return true
}
