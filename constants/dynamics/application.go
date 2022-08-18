package constants

import "math/big"

func bigFromBase10(s string) *big.Int {
	n, _ := new(big.Int).SetString(s, 10)
	return n
}

// DataStoreEpochFee is the initial fee for the DataStore object;
// this fee is the fee per epoch.
var DataStoreEpochFee = bigFromBase10("0")

// ValueStoreFee is the initial fee for the ValueStore object.
var ValueStoreFee = bigFromBase10("0")

// MinTxFee is the initial minimum transaction fee for a Tx object.
var MinTxFee = bigFromBase10("0")
