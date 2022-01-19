package dkg

import (
	"errors"
	"fmt"
	"math"
	"math/big"

	"github.com/MadBase/MadNet/blockchain/interfaces"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// RetrieveParticipants retrieves participant details from ETHDKG contract
func RetrieveParticipants(callOpts *bind.CallOpts, eth interfaces.Ethereum, logger *logrus.Entry) (objects.ParticipantList, int, error) {
	c := eth.Contracts()
	myIndex := math.MaxInt32

	addresses, err := c.ValidatorPool().GetValidatorAddresses(callOpts)
	if err != nil {
		message := fmt.Sprintf("could not get validator addresses from ValidatorPool: %v", err)
		logger.Errorf(message)
		return nil, myIndex, err
	}

	validatorStates, err := c.Ethdkg().GetParticipantsInternalState(callOpts, addresses)
	if err != nil {
		message := fmt.Sprintf("could not get internal states from Ethdkg: %v", err)
		logger.Errorf(message)
		return nil, myIndex, err
	}

	n := len(addresses)
	m := n

	// Now we process participant details
	participants := make(objects.ParticipantList, n)
	for i := 0; i < n; i++ {
		participantState := validatorStates[i]

		// todo: skip if participantState.Address == 0
		// because it means this validator is in the ValidatorPool
		// but has never registered in ETHDKG
		// if participantState.

		// Make corresponding Participant object
		participant := &objects.Participant{}
		participant.Address = addresses[i]
		participant.PublicKey = participantState.PublicKey

		if participantState.Index == 0 {
			participant.Index = m
			m--
		} else {
			participant.Index = int(participantState.Index)
		}

		participant.Nonce = participantState.Nonce
		participant.Phase = participantState.Phase
		participant.DistributedSharesHash = participantState.DistributedSharesHash
		participant.CommitmentsFirstCoefficient = participantState.CommitmentsFirstCoefficient
		participant.KeyShares = participantState.KeyShares
		participant.Gpkj = participantState.Gpkj

		// Set own index
		if callOpts.From == addresses[i] {
			myIndex = participant.Index
		}

		participants[participant.Index-1] = participant
	}

	return participants, myIndex, nil
}

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
