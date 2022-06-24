package brontide

import (
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/alicenet/alicenet/types"
)

// ErrWrongChainID occurs in (peer|self)ChainIdentifierHandshake
// when remoteID and localID fail to agree.
var ErrWrongChainID = errors.New("remote peer sent wrong chain identifier")

// Verify that both peers are working on the same chain by
// having them cross compare their chain identifiers.
// This step MUST be done after an authenticated encrypted channel
// have been built.
func selfInitiatedChainIdentifierHandshake(conn net.Conn, chainID types.ChainIdentifier) error {
	if err := writeUint32(conn, uint32(chainID)); err != nil {
		return err
	}
	// make sure other node is on same chain as us
	remoteCid, err := readUint32(conn)
	if err != nil {
		return err
	}
	if uint32(chainID) != remoteCid {
		return fmt.Errorf("%s: wanted %v, got %v", ErrWrongChainID, chainID, remoteCid)
	}
	// clear out the deadline since we managed to get past the handshake
	return nil
}

func peerInitiatedChainIdentifierHandshake(conn net.Conn, chainID types.ChainIdentifier) error {
	remoteCid, err := readUint32(conn)
	if err != nil {
		return err
	}
	if err := writeUint32(conn, uint32(chainID)); err != nil {
		return err
	}
	if uint32(chainID) != remoteCid {
		return fmt.Errorf("%s: wanted %v, got %v", ErrWrongChainID, chainID, remoteCid)
	}
	return nil
}

func selfInitiatedPortHandshake(conn net.Conn, port int) (int, error) {
	// cast can not overflow
	if err := writeUint32(conn, uint32(port)); err != nil {
		return 0, err
	}
	remotePort, err := readUint32(conn)
	if err != nil {
		return 0, err
	}
	return int(remotePort), nil
}

func peerInitiatedPortHandshake(conn net.Conn, port int) (int, error) {
	remotePort, err := readUint32(conn)
	if err != nil {
		return 0, err
	}
	if err := writeUint32(conn, uint32(port)); err != nil {
		return 0, err
	}
	return int(remotePort), nil
}

func selfInitiatedVersionHandshake(conn net.Conn, version types.ProtoVersion) (uint32, error) {
	if err := writeUint32(conn, uint32(version)); err != nil {
		return 0, err
	}
	remoteversion, err := readUint32(conn)
	if err != nil {
		return 0, err
	}
	return remoteversion, nil
}

func peerInitiatedVersionHandshake(conn net.Conn, version types.ProtoVersion) (uint32, error) {
	remoteversion, err := readUint32(conn)
	if err != nil {
		return 0, err
	}
	if err := writeUint32(conn, uint32(version)); err != nil {
		return 0, err
	}
	return remoteversion, nil
}

func writeUint32(conn net.Conn, local uint32) error {
	localBytes := marshalUint32(local)
	_, err := conn.Write(localBytes[:])
	if err != nil {
		return err
	}
	return nil
}

func readUint32(conn net.Conn) (uint32, error) {
	remotePortByteSlice := make([]byte, 4)
	_, err := io.ReadFull(conn, remotePortByteSlice)
	if err != nil {
		return 0, err
	}
	var remotePortBytes [4]byte
	copy(remotePortBytes[:], remotePortByteSlice)
	r := unmarshalUint32(remotePortBytes)
	return r, nil
}
