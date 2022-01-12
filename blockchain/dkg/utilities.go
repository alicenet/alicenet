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
func RetrieveParticipants(callOpts *bind.CallOpts, eth interfaces.Ethereum) (objects.ParticipantList, int, error) {
	c := eth.Contracts()
	myIndex := math.MaxInt32

	// Need to find how many participants there will be
	bigN, err := c.ValidatorPool().GetValidatorsCount(callOpts)
	if err != nil {
		return nil, myIndex, err
	}
	n := int(bigN.Uint64())

	// Now we retrieve participant details
	participants := make(objects.ParticipantList, n)
	for idx := 0; idx < n; idx++ {
		// First retrieve the address
		addr, err := c.ValidatorPool().GetValidator(callOpts, big.NewInt(int64(idx)))
		if err != nil {
			return nil, myIndex, err
		}

		participantState, err := c.Ethdkg().GetParticipantInternalState(callOpts, addr)
		if err != nil {
			return nil, myIndex, objects.ErrCanNotContinue
		}

		publicKey := participantState.PublicKey

		// Make corresponding Participant object
		participant := &objects.Participant{}
		participant.Address = addr
		participant.PublicKey = publicKey
		participant.Index = participant.Index

		// Set own index
		if callOpts.From == addr {
			myIndex = participant.Index
		}

		participants[idx] = participant
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
