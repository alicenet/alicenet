package dkgtasks

import (
	"bytes"
	"context"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/utils"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/sirupsen/logrus"
)

// GeneralTaskShouldRetry is the general logic used to determine if a task should try again
// -- Process is
func GeneralTaskShouldRetry(ctx context.Context, acct accounts.Account, logger *logrus.Entry,
	eth interfaces.Ethereum, publicKey [2]*big.Int,
	expectedRegistrationEnd uint64, expectedLastBlock uint64) bool {

	result := internalGeneralTaskShouldRetry(ctx, logger, eth, acct, publicKey, expectedRegistrationEnd, expectedLastBlock)

	logger.Infof("GeneralTaskRetry(acct:%v, publicKey:%v, expectedRegistrationEnd:%v, expectedLastBlock:%v): %v",
		acct.Address.Hex(),
		FormatPublicKey(publicKey),
		expectedRegistrationEnd,
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
	eth interfaces.Ethereum, acct accounts.Account, publicKey [2]*big.Int,
	expectedRegistrationEnd uint64, expectedLastBlock uint64) bool {

	// Make sure we're in the right block range to continue
	currentBlock, err := eth.GetCurrentHeight(ctx)
	if err != nil {
		// This probably means an endpoint issue, so we have to try again
		logger.Warnf("could not check current height of chain: %v", err)
		return true
	}
	if currentBlock > expectedLastBlock {
		return false
	}

	// This essentially checks if ethdkg was restarted while this task was running
	callOpts := eth.GetCallOpts(ctx, acct)
	registrationEnd, err := eth.Contracts().Ethdkg().TREGISTRATIONEND(callOpts)
	if err != nil {
		logger.Warnf("could not check when registration should have ended: %v", err)
		return true
	}

	if registrationEnd.Uint64() != expectedRegistrationEnd {
		return false
	}

	// Check to see if we are already registered
	status, err := CheckRegistration(ctx, eth.Contracts().Ethdkg(), logger, callOpts, acct.Address, publicKey)
	if err != nil {
		// This probably means an endpoint issue, so we have to try again
		logger.Warnf("could not check if we're registered: %v", err)
		return true
	}

	// If we aren't registered correctly retry won't work -- Not true for registration
	if status == NoRegistration || status == BadRegistration {
		logger.Warnf("registration status: %v", status)
		return false
	}

	// TODO Any other general cases where we know retry won't work?

	// We won't loop forever because eventually currentBlock > expectedLastBlock
	return true
}
