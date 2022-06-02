package utils

import (
	"bytes"
	"errors"
	"fmt"
	"math/big"

	ethereumInterfaces "github.com/MadBase/MadNet/blockchain/ethereum/interfaces"
	"github.com/MadBase/MadNet/blockchain/executor/tasks/dkg/objects"
	"github.com/MadBase/MadNet/constants"
	"github.com/MadBase/MadNet/crypto"
	"github.com/MadBase/MadNet/crypto/bn256"
	"github.com/MadBase/MadNet/crypto/bn256/cloudflare"
	"github.com/MadBase/MadNet/utils"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sirupsen/logrus"
)

// RetrieveGroupPublicKey retrieves participant's group public key (gpkj) from ETHDKG contract
func RetrieveGroupPublicKey(callOpts *bind.CallOpts, eth ethereumInterfaces.IEthereum, addr common.Address) ([4]*big.Int, error) {
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
func GetValidatorAddressesFromPool(callOpts *bind.CallOpts, eth ethereumInterfaces.IEthereum, logger *logrus.Entry) ([]common.Address, error) {
	c := eth.Contracts()

	addresses, err := c.ValidatorPool().GetValidatorsAddresses(callOpts)
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

// CategorizeGroupSigners returns 0 based indices of honest participants, 0 based indices of dishonest participants
func CategorizeGroupSigners(publishedPublicKeys [][4]*big.Int, participants objects.ParticipantList, commitments [][][2]*big.Int) (objects.ParticipantList, objects.ParticipantList, objects.ParticipantList, error) {
	// Setup + sanity checks before starting
	n := len(participants)
	threshold := ThresholdForUserCount(n)

	good := objects.ParticipantList{}
	bad := objects.ParticipantList{}
	missing := objects.ParticipantList{}

	// len(publishedPublicKeys) must equal len(publishedSignatures) must equal len(participants)
	if n != len(publishedPublicKeys) || n != len(commitments) {
		return objects.ParticipantList{}, objects.ParticipantList{}, objects.ParticipantList{}, fmt.Errorf(
			"mismatched public keys (%v), participants (%v), commitments (%v)", len(publishedPublicKeys), n, len(commitments))
	}

	// Require each commitment has length threshold+1
	for k := 0; k < n; k++ {
		if len(commitments[k]) != threshold+1 {
			return objects.ParticipantList{}, objects.ParticipantList{}, objects.ParticipantList{}, fmt.Errorf(
				"invalid commitments: required (%v); actual (%v)", threshold+1, len(commitments[k]))
		}
	}

	// We need commitments.
	// 		For each participant, loop through and form gpkj* term.
	//		Perform a PairingCheck to ensure valid gpkj.
	//		If invalid, add to bad list.

	g1Base := new(cloudflare.G1).ScalarBaseMult(common.Big1)
	orderMinus1 := new(big.Int).Sub(cloudflare.Order, common.Big1)
	h2Neg := new(cloudflare.G2).ScalarBaseMult(orderMinus1)

	// commitments:
	//		First dimension is participant index;
	//		Second dimension is commitment number
	for idx := 0; idx < n; idx++ {
		// Loop through all participants to confirm each is valid
		participant := participants[idx]

		// If public key is all zeros, then no public key was submitted;
		// add to missing.
		big0 := big.NewInt(0)
		if (publishedPublicKeys[idx][0] == nil ||
			publishedPublicKeys[idx][1] == nil ||
			publishedPublicKeys[idx][2] == nil ||
			publishedPublicKeys[idx][3] == nil) || (publishedPublicKeys[idx][0].Cmp(big0) == 0 &&
			publishedPublicKeys[idx][1].Cmp(big0) == 0 &&
			publishedPublicKeys[idx][2].Cmp(big0) == 0 &&
			publishedPublicKeys[idx][3].Cmp(big0) == 0) {
			missing = append(missing, participant.Copy())
			continue
		}

		j := participant.Index // participant index
		jBig := big.NewInt(int64(j))

		tmp0 := new(cloudflare.G1)
		gpkj, err := bn256.BigIntArrayToG2(publishedPublicKeys[idx])
		if err != nil {
			return objects.ParticipantList{}, objects.ParticipantList{}, objects.ParticipantList{}, fmt.Errorf("error converting BigIntArray to G2: %v", err)
		}

		// Outer loop determines what needs to be exponentiated
		for polyDegreeIdx := 0; polyDegreeIdx <= threshold; polyDegreeIdx++ {
			tmp1 := new(cloudflare.G1)
			// Inner loop loops through participants
			for participantIdx := 0; participantIdx < n; participantIdx++ {
				tmp2Big := commitments[participantIdx][polyDegreeIdx]
				tmp2, err := bn256.BigIntArrayToG1(tmp2Big)
				if err != nil {
					return objects.ParticipantList{}, objects.ParticipantList{}, objects.ParticipantList{}, fmt.Errorf("error converting BigIntArray to G1: %v", err)
				}
				tmp1.Add(tmp1, tmp2)
			}
			polyDegreeIdxBig := big.NewInt(int64(polyDegreeIdx))
			exponent := new(big.Int).Exp(jBig, polyDegreeIdxBig, cloudflare.Order)
			tmp1.ScalarMult(tmp1, exponent)

			tmp0.Add(tmp0, tmp1)
		}

		gpkjStar := new(cloudflare.G1).Set(tmp0)
		validPair := cloudflare.PairingCheck([]*cloudflare.G1{gpkjStar, g1Base}, []*cloudflare.G2{h2Neg, gpkj})
		if validPair {
			good = append(good, participant.Copy())
		} else {
			bad = append(bad, participant.Copy())
		}
	}

	return good, bad, missing, nil
}
