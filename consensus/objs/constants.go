package objs

func ProposalSigDesignator() []byte {
	return []byte("Proposal")
}

func PreVoteSigDesignator() []byte {
	return []byte("PreVote")
}

func PreVoteNilSigDesignator() []byte {
	return []byte("PreVoteNil")
}

func PreCommitSigDesignator() []byte {
	return []byte("PreCommit")
}

func PreCommitNilSigDesignator() []byte {
	return []byte("PreCommitNil")
}

func NextHeightSigDesignator() []byte {
	return []byte("NextHeight")
}

func NextRoundSigDesignator() []byte {
	return []byte("NextRound")
}
