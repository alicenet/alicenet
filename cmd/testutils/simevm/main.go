package main

import (
	"crypto/ecdsa"
	"math"
	"math/big"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alicenet/alicenet/blockchain"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	privateKeys := make([]*ecdsa.PrivateKey, 0)
	pk, err := crypto.HexToECDSA("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")

	if err != nil {
		panic(err)
	}

	privateKeys = append(privateKeys, pk)

	eth, err := blockchain.NewEthereumSimulator(
		privateKeys,
		6,
		1*time.Second,
		5*time.Second,
		0,
		big.NewInt(math.MaxInt64),
		50,
		math.MaxInt64,
		5*time.Second,
		30*time.Second)

	if err != nil {
		panic(err)
	}

	defer eth.Close()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals
}
