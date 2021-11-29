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
	bigN, err := c.Ethdkg().NumberOfRegistrations(callOpts)
	if err != nil {
		return nil, myIndex, err
	}
	n := int(bigN.Uint64())

	// Now we retrieve participant details
	participants := make(objects.ParticipantList, n)
	for idx := 0; idx < n; idx++ {
		// First retrieve the address
		addr, err := c.Ethdkg().Addresses(callOpts, big.NewInt(int64(idx)))
		if err != nil {
			return nil, myIndex, err
		}

		// Now the public keys
		var publicKey [2]*big.Int
		publicKey[0], err = c.Ethdkg().PublicKeys(callOpts, addr, common.Big0)
		if err != nil {
			return nil, myIndex, objects.ErrCanNotContinue
		}
		publicKey[1], err = c.Ethdkg().PublicKeys(callOpts, addr, common.Big1)
		if err != nil {
			return nil, myIndex, objects.ErrCanNotContinue
		}

		// Make corresponding Participant object
		participant := &objects.Participant{}
		participant.Address = addr
		participant.PublicKey = publicKey
		participant.Index = idx + 1

		// Set own index
		if callOpts.From == addr {
			myIndex = idx + 1
		}

		participants[idx] = participant
	}

	return participants, myIndex, nil
}

// RetrieveSignature retrieves participant's signature from ETHDKG contract
func RetrieveSignature(callOpts *bind.CallOpts, eth interfaces.Ethereum, addr common.Address) ([2]*big.Int, error) {
	var err error
	var sigBig [2]*big.Int

	ethdkg := eth.Contracts().Ethdkg()

	sigBig[0], err = ethdkg.InitialSignatures(callOpts, addr, common.Big0)
	if err != nil {
		return sigBig, err
	}

	sigBig[1], err = ethdkg.InitialSignatures(callOpts, addr, common.Big1)
	if err != nil {
		return sigBig, err
	}

	return sigBig, nil
}

// RetrieveGroupPublicKey retrieves participant's group public key (gpkj) from ETHDKG contract
func RetrieveGroupPublicKey(callOpts *bind.CallOpts, eth interfaces.Ethereum, addr common.Address) ([4]*big.Int, error) {
	var err error
	var gpkjBig [4]*big.Int

	ethdkg := eth.Contracts().Ethdkg()

	gpkjBig[0], err = ethdkg.GpkjSubmissions(callOpts, addr, common.Big0)
	if err != nil {
		return gpkjBig, err
	}

	gpkjBig[1], err = ethdkg.GpkjSubmissions(callOpts, addr, common.Big1)
	if err != nil {
		return gpkjBig, err
	}

	gpkjBig[2], err = ethdkg.GpkjSubmissions(callOpts, addr, common.Big2)
	if err != nil {
		return gpkjBig, err
	}

	gpkjBig[3], err = ethdkg.GpkjSubmissions(callOpts, addr, common.Big3)
	if err != nil {
		return gpkjBig, err
	}

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
