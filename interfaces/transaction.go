package interfaces

// Transaction is the minimum interface a transaction must meet for
// the consensus algorithm to work.
type Transaction interface {
	TxHash() ([]byte, error)
	MarshalBinary() ([]byte, error)
	XXXIsTx()
}
