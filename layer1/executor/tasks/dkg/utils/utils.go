package utils

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"

	"github.com/alicenet/alicenet/consensus/db"
	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/monitor/objects"
	"github.com/alicenet/alicenet/utils"
)

// RetrieveGroupPublicKey retrieves participant's group public key (gpkj) from ETHDKG contract.
func RetrieveGroupPublicKey(callOpts *bind.CallOpts, eth layer1.Client, contracts layer1.AllSmartContracts, addr common.Address) ([4]*big.Int, error) {
	var err error
	var gpkjBig [4]*big.Int

	ethdkg := contracts.EthereumContracts().Ethdkg()

	participantState, err := ethdkg.GetParticipantInternalState(callOpts, addr)
	if err != nil {
		return gpkjBig, err
	}

	gpkjBig = participantState.Gpkj

	return gpkjBig, nil
}

// IntsToBigInts converts an array of ints to an array of big ints.
func IntsToBigInts(ints []int) []*big.Int {
	bi := make([]*big.Int, len(ints))
	for idx, num := range ints {
		bi[idx] = big.NewInt(int64(num))
	}
	return bi
}

// LogReturnErrorf returns a formatted error for logger.
func LogReturnErrorf(logger *logrus.Entry, mess string, args ...interface{}) error {
	message := fmt.Sprintf(mess, args...)
	logger.Error(message)
	return errors.New(message)
}

// FormatPublicKey formats the public key suitably for logging.
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

// FormatBigIntSlice formats a slice of *big.Int's suitably for logging.
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

// GetValidatorAddresses retrieves validator addresses from the last monitor State saved on disk.
func GetValidatorAddresses(monitorDB *db.Database, logger *logrus.Entry) ([]common.Address, error) {
	monState, err := objects.GetMonitorState(monitorDB)
	if err != nil {
		return nil, fmt.Errorf("failed to get monitor state: %v", err)
	}
	var validatorAddresses []common.Address
	for address := range monState.PotentialValidators {
		validatorAddresses = append(validatorAddresses, address)
	}
	return validatorAddresses, nil
}

// GetValidatorAddresses retrieves validator addresses from the last monitor
// State saved on disk and check if a address sent is a potential validator
// address.
func IsValidator(monitorDB *db.Database, logger *logrus.Entry, address common.Address) (bool, error) {
	monState, err := objects.GetMonitorState(monitorDB)
	if err != nil {
		return false, fmt.Errorf("failed to get monitor state: %v", err)
	}
	_, present := monState.PotentialValidators[address]
	return present, nil
}
