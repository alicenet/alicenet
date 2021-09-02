package errorz

import (
	"errors"
	"fmt"
)

var (
	// ErrMissingTransactions should be raised by an IsValidFunc if a proposal may
	// not be verified due to missing transactions.
	ErrMissingTransactions = errors.New("unable to verify: missing transactions")
	ErrBadResponse         = errors.New("bad response from p2p request to remote peer")
	ErrClosing             = errors.New("shutting down, halt actions")
	ErrCorrupt             = errors.New("something went wrong that requires shutdown")
)

type ErrInvalid struct {
	msg     string
	wrapped error
}

func WrapErrInvalid(err error, fmtPattern string, fmtArgs ...interface{}) *ErrInvalid {
	if err == nil {
		return nil
	}

	return &ErrInvalid{msg: fmt.Sprintf(fmtPattern, fmtArgs...), wrapped: err}
}

func (e *ErrInvalid) Error() string {
	s := "the object is invalid:" + e.msg
	if e.wrapped != nil {
		return s + ":\n" + e.wrapped.Error()
	}
	return s
}

func (s *ErrInvalid) Unwrap() error {
	return s.wrapped
}

func (e ErrInvalid) New(msg string) *ErrInvalid {
	return &ErrInvalid{msg, nil}
}

type ErrConsensus struct {
	msg     string
	isLocal bool
}

func NewErrConsensus(msg string, isLocal bool) *ErrConsensus {
	return &ErrConsensus{msg: msg, isLocal: isLocal}
}

func (e *ErrConsensus) Error() string {
	return e.msg
}

type ErrStale struct {
	msg string
}

func (e *ErrStale) Error() string {
	return "the object is invalid:" + e.msg
}

func (e ErrStale) New(msg string, v ...interface{}) *ErrStale {
	return &ErrStale{fmt.Sprintf(msg, v...)}
}
