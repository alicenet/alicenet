package interfaces

// Lockable allows a mutex to be shared by interface.
type Lockable interface {
	Lock()
	Unlock()
}
