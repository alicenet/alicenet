package constants

import "time"

// Consensus params
const (
	DEADBLOCKROUND   uint32 = 5
	DEADBLOCKROUNDNR        = DEADBLOCKROUND - 1
	SrvrMsgTimeout          = 3 * time.Second // Do not go lower than 2 seconds!
	MsgTimeout              = 4 * time.Second // Do not go lower than 2 seconds!
)

// AdminHandlerKid returns a constant byte slice to be used as Key ID
func AdminHandlerKid() []byte {
	return []byte("constant")
}
