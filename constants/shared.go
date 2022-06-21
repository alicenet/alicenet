package constants

import "time"

const (
	// EpochLength is the number of blocks in an epoch for AliceNet
	EpochLength uint32 = 1024

	// HashLen specifies the length of a hash in bytes
	HashLen = 32

	// MaxUint32 is 2^32-1 for use as a indicator value
	MaxUint32 uint32 = 4294967295

	// MaxUint64 is 2^64-1
	MaxUint64 uint64 = 18446744073709551615

	// ETHDKGDesperationDelay is after how many Etereum blocks more validators will start being allowed. Highly dependent on EpochLength
	ETHDKGDesperationDelay int = 8

	// ETHDKGDesperationFactor determines how quickly more
	// validators will be allowed to perform unique ETHDKG
	// actions on MPKSubmission and Completion phases.
	// The lower this factor is, the quicker more validators
	// are elected as leaders.
	ETHDKGDesperationFactor int = 8
)

// CurveSpec specifies the particular elliptic curve we are dealing with
type CurveSpec uint8

const (
	// CurveSecp256k1 is the constant which specifies the curve Secp256k1;
	// this is the curve used by Ethereum
	CurveSecp256k1 CurveSpec = iota + 1

	// CurveBN256Eth is the constant which specifies the curve BN256;
	// this is the curve used in our crypto library for pairing-based crypto
	CurveBN256Eth
)

const (
	// CurveBN256EthPubkeyLen specifies the length of the public key for the
	// curve BN256; this is the uncompressed form
	CurveBN256EthPubkeyLen = 128
)

const (
	// CurveSecp256k1SigLen is the length of a Secp256k1 recoverable-ECDSA
	// digital signature
	CurveSecp256k1SigLen int = 65

	// CurveBN256EthSigLen is the length of a BN256 digital signature
	CurveBN256EthSigLen = 192
)

const (
	// OwnerLen is the constant which specifies the length of accounts
	// in bytes
	OwnerLen int = 20
)

// Status log keys
const (
	StatusLogger    = "status"
	StatusBlkTime   = "BlkTime"
	StatusGRCnt     = "GRCnt"
	StatusBlkRnd    = "Blk/Rnd"
	StatusBlkHsh    = "BlkHsh"
	StatusTxCt      = "TxCt"
	StatusSyncToBlk = "SyncToBlk"
)

// Logger names
const (
	LoggerConsensus = "consensus"
	LoggerTransport = "transport"
	LoggerApp       = "app"
	LoggerDB        = "db"
	LoggerGossipBus = "gossipbus"
	LoggerBadger    = "badger"
	LoggerPeerMan   = "peerMan"
	LoggerLocalRPC  = "localRPC"
	LoggerIPC       = "ipc"
	LoggerFirewalld = "firewalld"
	LoggerDMan      = "dman"
	LoggerPeer      = "peer"
	LoggerYamux     = "yamux"
	LoggerUPnP      = "upnp"
)

// Badger VLog GC ratio
const (
	BadgerDiscardRatio = 0.5
	MonDBGCFreq        = time.Duration(600)
)

// TODO Find a way to store this list that feels right
var ValidLoggers []string = []string{"alicenet", "consensus", "transport", "app", "db",
	"gossipbus", "badger", "peerman", "localrpc", "dman", "peer", "yamux",
	"ethereum", "main", "deploy", "utils", "monitor", "dkg",
	"services", "settings", "validator", "muxhandler", "bootnode", "p2pmux",
	"status", "test", "ipc", "firewalld"}
