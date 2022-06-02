package dbprefix

// All functions in this file are prefix designators for database state types.
// These functions name the resource being referenced in the function name.
// All prefixes should use two character length identifiers and should start
// at `a` as the first character allowed at index zero of an identifier.
// The identifiers should increase alpha-numeric from that point forward.

func PrefixOwnValidatingState() []byte {
	return []byte("aa")
}

func PrefixCurrentRoundState() []byte {
	return []byte("ab")
}

func PrefixHistoricRoundState() []byte {
	return []byte("ac")
}

func PrefixEncryptedStore() []byte {
	return []byte("ad")
}

func PrefixOwnState() []byte {
	return []byte("ae")
}

func PrefixValidatorSet() []byte {
	return []byte("af")
}

func PrefixBroadcastRCert() []byte {
	return []byte("ag")
}

func PrefixBroadcastProposal() []byte {
	return []byte("ah")
}

func PrefixBroadcastPreVote() []byte {
	return []byte("ai")
}

func PrefixBroadcastPreVoteNil() []byte {
	return []byte("aj")
}

func PrefixBroadcastPreCommit() []byte {
	return []byte("ak")
}

func PrefixBroadcastPreCommitNil() []byte {
	return []byte("al")
}

func PrefixBroadcastNextRound() []byte {
	return []byte("am")
}

func PrefixBroadcastNextHeight() []byte {
	return []byte("an")
}

func PrefixBroadcastBlockHeader() []byte {
	return []byte("ao")
}

func PrefixCommittedBlockHeader() []byte {
	return []byte("ap")
}

func PrefixCommittedBlockHeaderHashIndex() []byte {
	return []byte("aq")
}

func PrefixBlockHeaderTrie() []byte {
	return []byte("ar")
}

func PrefixBlockHeaderTrieRootCurrent() []byte {
	return []byte("as")
}

func PrefixBlockHeaderTrieRootHistoric() []byte {
	return []byte("at")
}

func PrefixSafeToProceed() []byte {
	return []byte("au")
}

func PrefixBroadcastTransaction() []byte {
	return []byte("av")
}

func PrefixSnapshotBlockHeader() []byte {
	return []byte("aw")
}

func PrefixTxCache() []byte {
	return []byte("ax")
}

func PrefixPendingNodeKey() []byte {
	return []byte("ay")
}

func PrefixPendingLeafKey() []byte {
	return []byte("az")
}

func PrefixPendingHdrNodeKey() []byte {
	return []byte("a1")
}

func PrefixPendingHdrLeafKey() []byte {
	return []byte("a2")
}

func PrefixStagedBlockHeaderKey() []byte {
	return []byte("a3")
}

func PrefixRawStorageKey() []byte {
	return []byte("a4")
}

func PrefixStorageNodeKey() []byte {
	return []byte("a5")
}

func PrefixPendingNodeKeyCount() []byte {
	return []byte("Ay")
}

func PrefixPendingLeafKeyCount() []byte {
	return []byte("Az")
}

func PrefixPendingHdrNodeKeyCount() []byte {
	return []byte("A1")
}

func PrefixPendingHdrLeafKeyCount() []byte {
	return []byte("A2")
}

func PrefixCommittedBlockHeaderCount() []byte {
	return []byte("A3")
}

func PrefixValidatorSetPostApplication() []byte {
	return []byte("ZZ")
}
