package dbprefix

func PrefixTaskSchedulerState() []byte {
	return []byte("schedulerStateKey")
}

func PrefixMonitorState() []byte {
	return []byte("monitorStateKey")
}

func PrefixTransactionWatcherState() []byte {
	return []byte("transactionWatcherKey")
}

func PrefixEthDKGState() []byte {
	return []byte("ethDkgStateKey")
}

func PrefixSnapshotState() []byte {
	return []byte("snapshotStateKey")
}
