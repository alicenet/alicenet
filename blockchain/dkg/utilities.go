package dkg

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/alicenet/alicenet/blockchain/interfaces"
	"github.com/alicenet/alicenet/bridge/bindings"
	"github.com/alicenet/alicenet/constants"
	"github.com/alicenet/alicenet/crypto"
	"github.com/alicenet/alicenet/crypto/bn256"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/sirupsen/logrus"
)

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

func WaitConfirmations(txHash common.Hash, ctx context.Context, logger *logrus.Entry, eth interfaces.Ethereum) error {
	var done = false

	for !done {

		receipt, err := eth.GetGethClient().TransactionReceipt(ctx, txHash)

		if err != nil {
			logger.Errorf("waiting for receipt failed: %v", err)
			return err
		}

		if receipt == nil {
			return errors.New("receipt is nil")
		}

		// Check receipt to confirm we were successful
		if receipt.Status != uint64(1) {
			message := fmt.Sprintf("tx status (%v) indicates failure: %v", receipt.Status, receipt.Logs)
			logger.Error(message)
			return errors.New(message)
		}

		receiptBlock := receipt.BlockNumber.Uint64()

		// receipt was successful, let's wait for `nConfirmation` block confirmations
		currentHeight, err := eth.GetCurrentHeight(ctx)
		if err != nil {
			return LogReturnErrorf(logger, "could not get block height: %v", err)
		}

		if currentHeight >= receiptBlock+eth.GetFinalityDelay() {
			done = true
		}

		time.Sleep(5 * time.Second)

	}

	return nil
}

func IncreaseFeeAndTipCap(gasFeeCap, gasTipCap *big.Int, percentage int, threshold uint64) (*big.Int, *big.Int) {
	// calculate percentage% increase in GasFeeCap
	var gasFeeCapPercent = (&big.Int{}).Mul(gasFeeCap, big.NewInt(int64(percentage)))
	gasFeeCapPercent = (&big.Int{}).Div(gasFeeCapPercent, big.NewInt(100))
	resultFeeCap := (&big.Int{}).Add(gasFeeCap, gasFeeCapPercent)

	// calculate percentage% increase in GasTipCap
	var gasTipCapPercent = (&big.Int{}).Mul(gasTipCap, big.NewInt(int64(percentage)))
	gasTipCapPercent = (&big.Int{}).Div(gasTipCapPercent, big.NewInt(100))
	resultTipCap := (&big.Int{}).Add(gasTipCap, gasTipCapPercent)

	if resultFeeCap.Uint64() > threshold {
		resultFeeCap = big.NewInt(int64(threshold))
	}

	return resultFeeCap, resultTipCap
}

func AmILeading(numValidators int, myIdx int, blocksSinceDesperation int, blockhash []byte, logger *logrus.Entry) bool {
	var numValidatorsAllowed int = 1
	for i := int(blocksSinceDesperation); i > 0; {
		i -= constants.ETHDKGDesperationFactor / numValidatorsAllowed
		numValidatorsAllowed++

		if numValidatorsAllowed >= numValidators {
			break
		}
	}

	// use the random nature of blockhash to deterministically define the range of validators that are allowed to take an ETHDKG action
	rand := (&big.Int{}).SetBytes(blockhash)
	start := int((&big.Int{}).Mod(rand, big.NewInt(int64(numValidators))).Int64())
	end := (start + numValidatorsAllowed) % numValidators

	if end > start {
		return myIdx >= start && myIdx < end
	} else {
		return myIdx >= start || myIdx < end
	}
}

func SetETHDKGPhaseLength(length uint16, eth interfaces.Ethereum, callOpts *bind.TransactOpts, ctx context.Context) (*types.Transaction, *types.Receipt, error) {
	// Shorten ethdkg phase for testing purposes
	ethdkgABI, err := abi.JSON(strings.NewReader(bindings.ETHDKGMetaData.ABI))
	if err != nil {
		return nil, nil, err
	}

	input, err := ethdkgABI.Pack("setPhaseLength", uint16(length))
	if err != nil {
		return nil, nil, err
	}

	txn, err := eth.Contracts().ContractFactory().CallAny(callOpts, eth.Contracts().EthdkgAddress(), big.NewInt(0), input)
	if err != nil {
		return nil, nil, err
	}
	if txn == nil {
		return nil, nil, errors.New("non existent transaction ContractFactory.CallAny(ethdkg, setPhaseLength(...))")
	}

	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		return nil, nil, err
	}
	if rcpt == nil {
		return nil, nil, errors.New("non existent receipt for tx ContractFactory.CallAny(ethdkg, setPhaseLength(...))")
	}

	return txn, rcpt, nil
}

func InitializeETHDKG(eth interfaces.Ethereum, callOpts *bind.TransactOpts, ctx context.Context) (*types.Transaction, *types.Receipt, error) {
	// Shorten ethdkg phase for testing purposes
	validatorPoolABI, err := abi.JSON(strings.NewReader(bindings.ValidatorPoolMetaData.ABI))
	if err != nil {
		return nil, nil, err
	}

	input, err := validatorPoolABI.Pack("initializeETHDKG")
	if err != nil {
		return nil, nil, err
	}

	txn, err := eth.Contracts().ContractFactory().CallAny(callOpts, eth.Contracts().ValidatorPoolAddress(), big.NewInt(0), input)
	if err != nil {
		return nil, nil, err
	}
	if txn == nil {
		return nil, nil, errors.New("non existent transaction ContractFactory.CallAny(validatorPool, initializeETHDKG())")
	}

	rcpt, err := eth.Queue().QueueAndWait(ctx, txn)
	if err != nil {
		return nil, nil, err
	}
	if rcpt == nil {
		return nil, nil, errors.New("non existent receipt for tx ContractFactory.CallAny(validatorPool, initializeETHDKG())")
	}

	return txn, rcpt, nil
}
