package handlers

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/chains/ethereum"
	"github.com/alicenet/alicenet/layer1/evm"
)

var _ layer1.AllSmartContracts = &AllSmartContractsHandle{}

// AllSmartContractsHandle is bus where we can access all smart contracts from the
// different layer1 clients.
type AllSmartContractsHandle struct {
	ethereumContracts *ethereum.Contracts
}

func NewAllSmartContractsHandle(
	eth *evm.Client,
	contractFactoryAddress common.Address,
) layer1.AllSmartContracts {
	return &AllSmartContractsHandle{
		ethereumContracts: ethereum.NewContracts(eth, contractFactoryAddress),
	}
}

func (ch *AllSmartContractsHandle) EthereumContracts() layer1.EthereumContracts {
	return ch.ethereumContracts
}
