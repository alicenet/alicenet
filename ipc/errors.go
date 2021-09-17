package ipc

import "fmt"

var (
	ErrNotSubscribed = fmt.Errorf("not subscribed")
	ErrNoConnection  = fmt.Errorf("no connection")
)
