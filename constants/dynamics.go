package constants

import (
	"math/big"
	"time"
)

const (
	InitialMaxBlockSize     uint64        = 3000000
	InitialProposalTimeout  time.Duration = 4 * time.Second
	InitialPreVoteTimeout   time.Duration = 3 * time.Second
	InitialPreCommitTimeout time.Duration = 3 * time.Second
)

// InitialDataStoreFee is the initial fee for the DataStore object;
// this fee is the fee per epoch.
var InitialDataStoreFee = new(big.Int).SetInt64(0)

// InitialValueStoreFee is the initial fee for the ValueStore object.
var InitialValueStoreFee = new(big.Int).SetInt64(0)

// InitialMinScaledTransactionFee is the initial minimum transaction fee for a Tx object.
var InitialMinScaledTransactionFee = new(big.Int).SetInt64(0)
