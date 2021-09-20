package errorz

import "fmt"

type ErrStale struct {
	*Err
}

func NewErrStale(msg string, v ...interface{}) *ErrStale {
	return &ErrStale{Err: NewErr(fmt.Sprintf(msg, v...))}
}

func (e *ErrStale) Error() string {
	return "the object is stale: " + e.Err.Error()
}

func (e *ErrStale) Wrap(err error) *ErrStale {
	e.Err.Wrap(err) // call method of embedded Err
	return e        // but return own reference to enable chaining
}

func (e *ErrStale) Trace(i ...interface{}) *ErrStale {
	e.Err.trace(1, i...) // call method of embedded Err
	return e             // but return own reference to enable chaining
}

func (e ErrStale) New(msg string, v ...interface{}) *ErrStale {
	// Defined for backwards compatibility, ideally update all dependent code and remove
	return NewErrStale(msg, v...)
}
