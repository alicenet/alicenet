package constants

// GRPC Server Configuration Params
// Setup to provide backpressure
const (
	ConsensusMsgQSize       = 1024
	ConsensusMsgQWorkers    = 4
	TxMsgQSize              = 1024
	TxMsgQWorkers           = 4
	LocalRPCMaxWorkers      = 2
	MaxConcurrentStreams    = 1
	P2PMaxConcurrentStreams = 4
	ReadBufferSize          = 0
	P2PStreamWorkers        = 4
	DiscoStreamWorkers      = 1
)
