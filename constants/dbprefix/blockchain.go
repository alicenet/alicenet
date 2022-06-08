package dbprefix

func PrefixTaskSchedulerState() []byte {
	return []byte("schedulerStateKey")
}

func PrefixMonitorState() []byte {
	return []byte("monitorStateKey")
}

func PrefixEthDKGState() []byte {
	return []byte("ethDkgStateKey")
}

func PrefixSnapshotState() []byte {
	return []byte("snapshotStateKey")
}
