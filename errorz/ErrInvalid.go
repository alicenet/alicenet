package errorz

type ErrInvalid struct {
	*Err
}

func NewErrInvalid(msg string) *ErrInvalid {
	return &ErrInvalid{Err: NewErr(msg)}
}

func (e *ErrInvalid) Error() string {
	return "the object is invalid: " + e.Err.Error()
}

//nolint:errcheck
func (e *ErrInvalid) Wrap(err error) *ErrInvalid {
	e.Err.Wrap(err) // call method of embedded Err
	return e        // but return own reference to enable chaining
}

//nolint:errcheck
func (e *ErrInvalid) Trace(i ...interface{}) *ErrInvalid {
	e.Err.trace(1, i...) // call method of embedded Err
	return e             // but return own reference to enable chaining
}

func (e ErrInvalid) New(msg string) *ErrInvalid {
	// Defined for backwards compatibility, ideally update all dependent code and remove
	return NewErrInvalid(msg)
}
