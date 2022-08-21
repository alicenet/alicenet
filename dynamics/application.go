package dynamics

import "math/big"

func bigFromBase10(s string) *big.Int {
	n, _ := new(big.Int).SetString(s, 10)
	return n
}

// dataStoreEpochFee is the initial fee for the DataStore object;
// this fee is the fee per epoch.
var dataStoreEpochFee = bigFromBase10("0")

// valueStoreFee is the initial fee for the ValueStore object.
var valueStoreFee = bigFromBase10("0")

// minTxFee is the initial minimum transaction fee for a Tx object.
var minTxFee = bigFromBase10("0")
