package dkg

import (
	"errors"
	"fmt"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

/*
// RetrieveParticipants retrieves participant details from ETHDKG contract
func RetrieveParticipants(participants []common.Address, callOpts *bind.CallOpts, eth interfaces.Ethereum, logger *logrus.Entry) (objects.ParticipantList, int, error) {
	c := eth.Contracts()
	myIndex := math.MaxInt32

	// addresses, err := c.ValidatorPool().GetValidatorAddresses(callOpts)
	// if err != nil {
	// 	message := fmt.Sprintf("could not get validator addresses from ValidatorPool: %v", err)
	// 	logger.Errorf(message)
	// 	return nil, myIndex, err
	// }

	validatorStates, err := c.Ethdkg().GetParticipantsInternalState(callOpts, participants)
	if err != nil {
		message := fmt.Sprintf("could not get internal states from Ethdkg: %v", err)
		logger.Errorf(message)
		return nil, myIndex, err
	}

	// nValidators, err := c.Ethdkg().GetNumValidators(callOpts)
	// if err != nil {
	// 	message := fmt.Sprintf("could not get number of validators from Ethdkg: %v", err)
	// 	logger.Errorf(message)
	// 	return nil, myIndex, err
	// }

	var expectedNumValidators = len(participants)
	// var numValidators = len(participants)

	var n = expectedNumValidators

	// if numValidators > expectedNumValidators {
	// 	n = numValidators
	// }

	// m := n

	// Now we process participant details
	participantStates := make(objects.ParticipantList, n)
	for i := 0; i < n; i++ {
		participantState := validatorStates[i]

		// todo: skip if participantState.Address == 0
		// because it means this validator is in the ValidatorPool
		// but has never registered in ETHDKG
		// if participantState.

		// Make corresponding Participant object
		participant := &objects.Participant{}
		participant.Address = participants[i]
		participant.PublicKey = participantState.PublicKey

		// if participantState.Index == 0 {
		// 	participant.Index = m
		// 	m--
		// } else {
		// 	participant.Index = int(participantState.Index)
		// }

		participant.Nonce = participantState.Nonce
		participant.Phase = participantState.Phase
		participant.DistributedSharesHash = participantState.DistributedSharesHash
		participant.CommitmentsFirstCoefficient = participantState.CommitmentsFirstCoefficient
		participant.KeyShares = participantState.KeyShares
		participant.Gpkj = participantState.Gpkj

		// Set own index
		if callOpts.From == participants[i] {
			myIndex = participant.Index
		}

		participantStates[participant.Index-1] = participant
	}

	return participantStates, myIndex, nil
}
*/

// RetrieveGroupPublicKey retrieves participant's group public key (gpkj) from ETHDKG contract
func RetrieveGroupPublicKey(callOpts *bind.CallOpts, eth interfaces.Ethereum, addr common.Address) ([4]*big.Int, error) {
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

// GetValidatorAddressesFromPool retrieves validator addresses from ValidatorPool
func GetValidatorAddressesFromPool(callOpts *bind.CallOpts, eth interfaces.Ethereum, logger *logrus.Entry) ([]common.Address, error) {
	c := eth.Contracts()

	addresses, err := c.ValidatorPool().GetValidatorAddresses(callOpts)
	if err != nil {
		message := fmt.Sprintf("could not get validator addresses from ValidatorPool: %v", err)
		logger.Errorf(message)
		return nil, err
	}

	return addresses, nil
}

// ComputeDistributedSharesHash computes the distributed shares hash, encrypted shares hash and commitments hash
func ComputeDistributedSharesHash(encryptedShares []*big.Int, commitments [][2]*big.Int) ([32]byte, [32]byte, [32]byte, error) {
	var emptyBytes32 [32]byte

	// encrypted shares hash
	encryptedSharesBin, err := bn256.MarshalBigIntSlice(encryptedShares)
	if err != nil {
		return emptyBytes32, emptyBytes32, emptyBytes32, fmt.Errorf("ComputeDistributedSharesHash encryptedSharesBin failed: %v", err)
	}
	hashSlice := crypto.Hasher(encryptedSharesBin)
	var encryptedSharesHash [32]byte
	copy(encryptedSharesHash[:], hashSlice)

	// commitments hash
	commitmentsBin, err := bn256.MarshalG1BigSlice(commitments)
	if err != nil {
		return emptyBytes32, emptyBytes32, emptyBytes32, fmt.Errorf("ComputeDistributedSharesHash commitmentsBin failed: %v", err)
	}
	hashSlice = crypto.Hasher(commitmentsBin)
	var commitmentsHash [32]byte
	copy(commitmentsHash[:], hashSlice)

	// distributed shares hash
	var distributedSharesBin = append(encryptedSharesHash[:], commitmentsHash[:]...)
	hashSlice = crypto.Hasher(distributedSharesBin)
	var distributedSharesHash [32]byte
	copy(distributedSharesHash[:], hashSlice)

	return distributedSharesHash, encryptedSharesHash, commitmentsHash, nil
}
