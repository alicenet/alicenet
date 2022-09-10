package dynamics

// Ensuring interface check
var _ Updater = (*Update)(nil)

// Updater specifies the interface we use for updating Storage
type Updater interface {
	Value() []byte
	Epoch() uint32
}

// Update is an implementation of Updater interface
type Update struct {
	value []byte
	epoch uint32
}

// NewUpdate makes a valid valid Update struct which is then used
func NewUpdate(value []byte, epoch uint32) *Update {
	return &Update{
		value: value,
		epoch: epoch,
	}
}

// Value returns the value of Update
func (u *Update) Value() []byte {
	return u.value
}

// Epoch returns the epoch of Update
func (u *Update) Epoch() uint32 {
	return u.epoch
}
