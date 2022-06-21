package transport

import (
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/alicenet/alicenet/crypto/secp256k1"
	"github.com/alicenet/alicenet/interfaces"
	"github.com/alicenet/alicenet/transport/brontide"
	"github.com/alicenet/alicenet/types"
)

var _ net.Addr = (*NodeAddr)(nil)
var _ interfaces.NodeAddr = (*NodeAddr)(nil)
var _ error = (*ErrInvalidNetworkAddress)(nil)

// ErrInvalidNetworkAddress allows us to handle complicated errors
// where additional information is required.
type ErrInvalidNetworkAddress struct {
	msg string
}

func (eina *ErrInvalidNetworkAddress) Error() string {
	return eina.msg
}

// Constants used for parsing NodeAddr from a string.
const (
	addressSeparator                   string = "@"
	chainSeparator                     string = "|"
	compressedPublicKeyHexStringLength int    = 66
	networkString                      string = "tcp"
)

// NodeAddr is the peer to peer address type definition.
// This definition MUST comply with the NodeAddr interface.
type NodeAddr struct {
	host     string
	port     int
	identity *secp256k1.PublicKey
	chainID  types.ChainIdentifier
}

// Unmarshal will return a NEW NodeAddr by unmarshalling the string of the
// address into a NodeAddr object.
func (pad *NodeAddr) Unmarshal(address string) (interfaces.NodeAddr, error) {
	return NewNodeAddr(address)
}

// Network returns the network string as expected by a net.Addr
// interface would return it. This is part of the net.Addr
// interface.
func (pad *NodeAddr) Network() string {
	return networkString
}

// Returns the address string as expected by a net.Addr
// interface would return it. This is part of the net.Addr
// interface.
func (pad *NodeAddr) String() string {
	return net.JoinHostPort(pad.host, strconv.Itoa(pad.port))
}

// Identity returns the hex string representation of the public key of the
// remote peer. This is a unique identifier to each node.
func (pad *NodeAddr) Identity() string {
	return pubkeyToIdent(pad.identity)
}

// P2PAddr returns the address of the connection as a properly formatted
// string for use in dialing a peer.
func (pad *NodeAddr) P2PAddr() string {
	return formatNodeAddr(pad.chainID, pad.Identity(), pad.host, pad.port)
}

// ChainID returns the chainID of the address.
func (pad *NodeAddr) ChainID() types.ChainIdentifier {
	return pad.chainID
}

// Host returns the host of the NodeAddr
func (pad *NodeAddr) Host() string {
	return pad.host
}

// Port returns the port of the address.
func (pad *NodeAddr) Port() int {
	return pad.port
}

// ChainIdentifier identifies the chain this connection is expecting it's
// peers to also be a member of.
func (pad *NodeAddr) ChainIdentifier() types.ChainIdentifier {
	return pad.chainID
}

// ToBTCNetAddr converts the address into the format brontide expects.
// This function is not part of the interface definition of
// a NodeAddr so that these assumptions remain isolated to the
// transport package
func (pad *NodeAddr) toBTCNetAddr() *brontide.NetAddress {
	return &brontide.NetAddress{
		Address: &addr{
			network: pad.Network(),
			address: net.JoinHostPort(pad.host, strconv.Itoa(pad.port))},
		IdentityKey: pad.identity,
	}
}

// NewNodeAddr constructs a new NodeAddr given an address of valid format.
// The network should ALWAYS be `tcp` due to the use of brontide.
// The address MUST be formatted as follows:
// <8 hex characters>|<66 hex characters>@host:port
// <chainIdentifier>|<hex PublicKey>@host:port
func NewNodeAddr(address string) (interfaces.NodeAddr, error) {
	chainID, pubkeyhexstring, nodeAddr, err := splitNetworkAddress(address)
	if err != nil {
		return nil, err
	}
	if len(pubkeyhexstring) != compressedPublicKeyHexStringLength {
		return nil, ErrInvalidPubKeyLength
	}
	pubkeybytes, err := hex.DecodeString(pubkeyhexstring)
	if err != nil {
		return nil, err
	}
	pubkey, err := secp256k1.ParsePubKey(pubkeybytes, secp256k1.S256())
	if err != nil {
		return nil, err
	}
	thisNodeAddr := &addr{
		network: tcpNetwork,
		address: nodeAddr,
	}
	host, port, err := net.SplitHostPort(thisNodeAddr.String())
	if err != nil {
		return nil, err
	}
	ps, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}
	return &NodeAddr{
		host:     host,
		port:     ps,
		identity: pubkey,
		chainID:  types.ChainIdentifier(chainID),
	}, nil
}

// Serializes a public key to a lower case hex encoded string
func pubkeyToIdent(pubk *secp256k1.PublicKey) string {
	p := pubk.SerializeCompressed()
	return fmt.Sprintf("%x", p)
}

// Formats a node address given the proper inputs
func formatNodeAddr(cid types.ChainIdentifier, ident string, host string, port int) string {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	return fmt.Sprintf("%s%s%s%s%s", uint32ToHexString(uint32(cid)), chainSeparator, ident, addressSeparator, addr)
}

// RandomNodeAddr returns a random NodeAddr for peer discovery lookups
func RandomNodeAddr() (interfaces.NodeAddr, error) {
	privk, err := newTransportPrivateKey()
	if err != nil {
		return nil, err
	}
	pubkey := publicKeyFromPrivateKey(privk)
	return &NodeAddr{
		identity: pubkey,
	}, nil
}

// SplitNetworkAddress splits a network address into three parts
// returns them as <ChainIdentifier>, <hex encoded pubkey>, <address>, <discoveryPort>
// the last returned values of <address> may be split to host and port using
// net.SplitHostPort
// the first <address> is the discovery address, the second is the NodeAddr
func splitNetworkAddress(address string) (uint32, string, string, error) {
	parts := strings.Split(address, chainSeparator)
	if len(parts) != 2 {
		return 0, "", "", &ErrInvalidNetworkAddress{fmt.Sprintf("invalid network address: Chain Separator: %s", address)}
	}
	chainIdentifierHex := parts[0]
	chainID, err := uint32FromHexString(chainIdentifierHex)
	if err != nil {
		return 0, "", "", &ErrInvalidNetworkAddress{fmt.Sprintf("%s: %v", err.Error(), chainIdentifierHex)}
	}
	pubkeyAndAddress := parts[1]
	subParts := strings.Split(pubkeyAndAddress, addressSeparator)
	if len(subParts) != 2 {
		return 0, "", "", &ErrInvalidNetworkAddress{"invalid network address: Address Separator"}
	}
	if len(subParts[0]) != compressedPublicKeyHexStringLength {
		return 0, "", "", &ErrInvalidNetworkAddress{"invalid Node Identifier hex string"}
	}
	return chainID, subParts[0], subParts[1], nil
}
