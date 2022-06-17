package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/ethereum"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// RetrieveGroupPublicKey retrieves participant's group public key (gpkj) from ETHDKG contract
func RetrieveGroupPublicKey(callOpts *bind.CallOpts, eth ethereum.Network, addr common.Address) ([4]*big.Int, error) {
	var err error
	var gpkjBig [4]*big.Int

	ethdkg := eth.Contracts().Ethdkg()

	participantState, err := ethdkg.GetParticipantInternalState(callOpts, addr)
	if err != nil {
		return gpkjBig, err
	}

	gpkjBig = participantState.Gpkj

	return gpkjBig, nil
}

// IntsToBigInts converts an array of ints to an array of big ints
func IntsToBigInts(ints []int) []*big.Int {
	bi := make([]*big.Int, len(ints))
	for idx, num := range ints {
		bi[idx] = big.NewInt(int64(num))
	}
	return bi
}

// LogReturnErrorf returns a formatted error for logger
func LogReturnErrorf(logger *logrus.Entry, mess string, args ...interface{}) error {
	message := fmt.Sprintf(mess, args...)
	logger.Error(message)
	return errors.New(message)
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

// GetValidatorAddressesFromPool retrieves validator addresses from ValidatorPool
func GetValidatorAddressesFromPool(callOpts *bind.CallOpts, eth ethereum.Network, logger *logrus.Entry) ([]common.Address, error) {
	c := eth.Contracts()

	addresses, err := c.ValidatorPool().GetValidatorsAddresses(callOpts)
	if err != nil {
		message := fmt.Sprintf("could not get validator addresses from ValidatorPool: %v", err)
		logger.Errorf(message)
		return nil, err
	}

	return addresses, nil
}

// check if I'm a leader for this task
func AmILeading(client ethereum.Network, ctx context.Context, logger *logrus.Entry, start int, startBlockHash []byte, numValidators int, dkgIndex int) bool {
	currentHeight, err := client.GetCurrentHeight(ctx)
	if err != nil {
		return false
	}

	blocksSinceDesperation := int(currentHeight) - start - constants.ETHDKGDesperationDelay
	amILeading := utils.AmILeading(numberOfValidators, dkgIndex-1, blocksSinceDesperation, startBlockHash, logger)

	logger.WithFields(logrus.Fields{
		"currentHeight":                    currentHeight,
		"t.Start":                          start,
		"constants.ETHDKGDesperationDelay": constants.ETHDKGDesperationDelay,
		"blocksSinceDesperation":           blocksSinceDesperation,
		"amILeading":                       amILeading,
	}).Infof("dkg.AmILeading")

	return amILeading
}
