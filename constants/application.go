package constants

const (
	// MaxTxVectorLength is the maximum size of input output vectors.
	// This prevents uint32 overflow.
	MaxTxVectorLength int = 128
)

const (
	// DSPIMinDeposit is the minimum amount of deposit. This is calculated
	// assuming that no state is stored (datasize == 0) as well as storing
	// the state for 1 epoch.
	DSPIMinDeposit uint32 = BaseDatasizeConst

	// BaseDatasizeConst is the bytes added to the size of state (in bytes)
	// for the minimum cost.
	BaseDatasizeConst = 376
	// We discuss the rational now.
	//
	// At this point there are two possibilities for accounts,
	// Secp256k1 and BN256Eth. We have the following
	//
	//	Curve 			Size of RawData 	len(DataStore.MarshalBinary())
	//
	//	Secp256k1				32						272
	//	         				64						304
	//	         				96						336
	//
	//	Bn256Eth 				32						400
	//	         				64						432
	//	         				96						464
	//
	// From looking at this table, the base cost for Secp256k1 is 240 bytes
	// while the base cost for BN256Eth is 368 bytes. Many machines have
	// word size of 64 bits, so we add an additional 8 bytes to the baseline
	// to ensure all storage is needed to ensure proper alignment.
	// Thus, it makes sense to set 376 bytes as the base cost of DataStore.

	// MaxDataStoreSize is the largest size of RawData that we store in a
	// DataStore; 2 MiB (2^21)
	//
	MaxDataStoreSize uint32 = 2097152
	//
	// Do not change this value without ensuring BaseDepositEquation will
	// not overflow in uint64. This will not happen so long as
	//
	// 		MaxDataStoreSize + BaseDatasizeConst < 2^31
	//
	// This restriction should not cause problems.
)
