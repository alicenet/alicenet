package types

// ProtoVersion is a custom type used to store protocol version
type ProtoVersion uint32

// Protocol specifies if this is a P2P or a discovery connection.
type Protocol uint32

// These types designate the protocol a dialer is asking to bind against.
// If this value is P2PProtocol, a peer to peer protocol is being initiated.
// if this value is DiscProtocol, a discovery protocol is being initiated.
// If this value is Bootnode, a bootnode protocol is being initiated.
const (
	P2PProtocol = Protocol(iota + 1)
	DiscProtocol
	Bootnode
)
