package errorz

type ErrConsensus struct {
	*Err
	isLocal bool
}

func NewErrConsensus(msg string, isLocal bool) *ErrConsensus {
	return &ErrConsensus{Err: NewErr(msg), isLocal: isLocal}
}

func (e *ErrConsensus) Error() string {
	return "consensus error: " + e.Err.Error()
}

func (e *ErrConsensus) IsLocal() bool {
	return e.isLocal
}

func (e *ErrConsensus) Wrap(err error) *ErrConsensus {
	e.Err.Wrap(err) // call method of embedded Err
	return e        // but return own reference to enable chaining
}

func (e *ErrConsensus) Trace(i ...interface{}) *ErrConsensus {
	e.Err.trace(1, i...) // call method of embedded Err
	return e             // but return own reference to enable chaining
}
