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

type Err struct {
	msg     string
	wrapped error
	traces  []string
}

func NewErr(msg string) *Err {
	return &Err{msg, nil, []string{}}
}

// Unwrap returns the wrapper (inner) error
func (s *Err) Unwrap() error {
	return s.wrapped
}

// Wrap wraps this error around a given error
func (e *Err) Wrap(inner error) *Err {
	if inner == nil {
		return nil
	}

	e.wrapped = inner
	return e
}

// Trace adds a trace line to this error with an optional suffix
// Suffix can be a single value, or an format string followed by format values
func (e *Err) Trace(suffix ...interface{}) *Err {
	return e.trace(1, suffix...)
}

func (e *Err) trace(depth int, suffix ...interface{}) *Err {
	if e == nil {
		return nil
	}
	if e.traces == nil {
		e.traces = make([]string, 1)
	}

	trace := MakeTrace(depth + 1)

	if len(suffix) == 1 {
		e.traces = append(e.traces, fmt.Sprintf("%v %v", trace, suffix[0]))
	} else if len(suffix) > 1 {
		e.traces = append(e.traces, trace+" "+fmt.Sprintf(suffix[0].(string), suffix[1:]...))
	} else {
		e.traces = append(e.traces, trace)
	}

	return e
}

// Error returns the error message and traces of this error and any wrapped errors
func (e *Err) Error() string {
	ret := e.msg

	if len(e.traces) > 0 || e.wrapped != nil {
		ret += ":"
	}

	for _, v := range e.traces {
		ret += "\n\t" + v
	}

	if e.wrapped != nil {
		ret += "\n" + e.wrapped.Error()
	}

	return ret
}
