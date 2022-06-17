package utils

import (
	"math/big"

	"github.com/MadBase/MadNet/constants"
	"github.com/sirupsen/logrus"
)

func AmILeading(numValidators int, myIdx int, blocksSinceDesperation int, blockHash []byte, logger *logrus.Entry) bool {
	var numValidatorsAllowed int = 1
	for i := int(blocksSinceDesperation); i > 0; {
		i -= constants.ETHDKGDesperationFactor / numValidatorsAllowed
		numValidatorsAllowed++

		if numValidatorsAllowed >= numValidators {
			break
		}
	}

	// use the random nature of blockhash to deterministically define the range of validators that are allowed to take an ETHDKG action
	rand := (&big.Int{}).SetBytes(blockHash)
	start := int((&big.Int{}).Mod(rand, big.NewInt(int64(numValidators))).Int64())
	end := (start + numValidatorsAllowed) % numValidators

	if end > start {
		return myIdx >= start && myIdx < end
	} else {
		return myIdx >= start || myIdx < end
	}
}
