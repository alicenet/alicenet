package types

// P2PInitiator specifies who was the initiating party.
// If the remote peer dialed the local server, the value is
// PeerInitiatedConnection
// If the remote peer was dialed by the local client, the value is
// SelfInitiatedConnection.
type P2PInitiator uint8

// Specifies who was the initiating party for a connection.
// If the remote peer dialed the local server, the value is
// PeerInitiatedConnection
// If the remote peer was dialed by the local client, the value is
// SelfInitiatedConnection.
const (
	SelfInitiatedConnection = P2PInitiator(iota + 1)
	PeerInitiatedConnection
)
