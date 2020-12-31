package errorz

import "errors"

var (
	// ErrMissingTransactions should be raised by an IsValidFunc if a proposal may
	// not be verified due to missing transactions.
	ErrMissingTransactions = errors.New("unable to verify: missing transactions")
	ErrBadResponse         = errors.New("bad response from p2p request to remote peer")
	ErrClosing             = errors.New("shutting down, halt actions")
	ErrCorrupt             = errors.New("something went wrong that requires shutdown")
)

type ErrInvalid struct {
	msg string
}

func (e *ErrInvalid) Error() string {
	return "the object is invalid:" + e.msg
}

func (e ErrInvalid) New(msg string) *ErrInvalid {
	return &ErrInvalid{msg}
}

type ErrStale struct {
	msg string
}

func (e *ErrStale) Error() string {
	return "the object is invalid:" + e.msg
}

func (e ErrStale) New(msg string) *ErrStale {
	return &ErrStale{msg}
}
