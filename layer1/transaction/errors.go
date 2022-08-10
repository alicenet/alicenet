package transaction

// Error type that is usually recoverable (e.g network issues).
type ErrRecoverable struct {
	message string
}

func (e *ErrRecoverable) Error() string {
	return e.message
}

// Specific error in case a transaction becomes stale.
type ErrTransactionStale struct {
	message string
}

func (e *ErrTransactionStale) Error() string {
	return e.message
}

// Error in case a transaction is not found in the rpc node.
type ErrTxNotFound struct {
	message string
}

func (e *ErrTxNotFound) Error() string {
	return e.message
}

// Error in case of a invalid request is sent to the watcher backend.
type ErrInvalidMonitorRequest struct {
	message string
}

func (e *ErrInvalidMonitorRequest) Error() string {
	return e.message
}

// Error in case of a invalid transaction is sent to the watcher backend.
type ErrInvalidTransactionRequest struct {
	message string
}

func (e *ErrInvalidTransactionRequest) Error() string {
	return e.message
}
