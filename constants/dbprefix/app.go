package dbprefix

// All functions in this file are prefix designators for database state types.
// These functions name the resource being referenced in the function name.
// All prefixes should use two character length identifiers and should start
// at `n` as the first character allowed at index zero of an identifier.
// The identifiers should increase alpha-numeric from that point forward.

func PrefixMinedTx() []byte {
	return []byte("na")
}

func PrefixMinedTxIndexRefKey() []byte {
	return []byte("nb")
}

func PrefixMinedTxIndexKey() []byte {
	return []byte("nc")
}

func PrefixTrieRootForHeight() []byte {
	return []byte("nd")
}

func PrefixUTXOTrie() []byte {
	return []byte("nl")
}

func PrefixCurrentStateRoot() []byte {
	return []byte("nm")
}

func PrefixPendingStateRoot() []byte {
	return []byte("nn")
}

func PrefixCanonicalStateRoot() []byte {
	return []byte("no")
}

func PrefixPendingTx() []byte {
	return []byte("np")
}

func PrefixPendingTxEpochConstraintListRef() []byte {
	return []byte("nq")
}

func PrefixPendingTxEpochConstraintList() []byte {
	return []byte("nr")
}

func PrefixPendingTxInsertionOrderIndex() []byte {
	return []byte("ns")
}

func PrefixPendingTxInsertionOrderReverseIndex() []byte {
	return []byte("nt")
}

func PrefixMinedUTXO() []byte {
	return []byte("nu")
}

func PrefixMinedUTXOEpcKey() []byte {
	return []byte("nv")
}

func PrefixMinedUTXOEpcRefKey() []byte {
	return []byte("nw")
}

func PrefixMinedUTXODataKey() []byte {
	return []byte("nx")
}

func PrefixMinedUTXODataRefKey() []byte {
	return []byte("ny")
}

func PrefixMinedUTXOValueRefKey() []byte {
	return []byte("nz")
}

func PrefixMinedUTXOValueKey() []byte {
	return []byte("n0")
}

func PrefixUTXORefLinker() []byte {
	return []byte("n1")
}

func PrefixUTXORefLinkerRev() []byte {
	return []byte("n2")
}

func PrefixUTXOCounter() []byte {
	return []byte("n3")
}

func PrefixDeposit() []byte {
	return []byte("n4")
}

func PrefixDepositValueRefKey() []byte {
	return []byte("n5")
}

func PrefixDepositValueKey() []byte {
	return []byte("n6")
}

func PrefixPendingTxCooldownKey() []byte {
	return []byte("n7")
}
