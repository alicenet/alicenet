package dbprefix

// SERVICES
// All functions in this file are prefix designators for database state types.
// These functions name the resource being referenced in the function name.
// All prefixes should use two character length identifiers and should start
// at `l` as the first character allowed at index zero of an identifier.
// The identifiers should increase alpha-numeric from that point forward.
func PrefixTaskSchedulerState() []byte {
	return []byte("la")
}

func PrefixMonitorState() []byte {
	return []byte("lb")
}

func PrefixTransactionWatcherState() []byte {
	return []byte("lc")
}

// TASKS
// All functions in this file are prefix designators for database state types.
// These functions name the resource being referenced in the function name.
// All prefixes should use two character length identifiers and should start
// at `t` as the first character allowed at index zero of an identifier.
// The identifiers should increase alpha-numeric from that point forward.
func PrefixEthDKGState() []byte {
	return []byte("ta")
}

func PrefixSnapshotState() []byte {
	return []byte("tb")
}
