package dkg

import (
	"math"
	"math/big"

	"github.com/MadBase/MadNet/blockchain"
	"github.com/MadBase/MadNet/blockchain/objects"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// RetrieveParticipants retrieves participant details from ETHDKG contract
func RetrieveParticipants(eth blockchain.Ethereum, callOpts *bind.CallOpts) (objects.ParticipantList, int, error) {

	c := eth.Contracts()
	myIndex := math.MaxInt32

	// Need to find how many participants there will be
	bigN, err := c.Ethdkg().NumberOfRegistrations(callOpts)
	if err != nil {
		return nil, myIndex, err
	}
	n := int(bigN.Uint64())

	// Now we retrieve participant details
	participants := make(objects.ParticipantList, int(n))
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

		participant := &objects.Participant{}
		participant.Address = addr
		participant.PublicKey = publicKey
		participant.Index = idx

		if callOpts.From == addr {
			myIndex = idx
		}

		participants[idx] = participant
	}

	return participants, myIndex, nil
}
