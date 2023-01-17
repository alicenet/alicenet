package handlers

import (
	"github.com/ethereum/go-ethereum/common"

	"github.com/alicenet/alicenet/layer1"
	"github.com/alicenet/alicenet/layer1/chains/ethereum"
	"github.com/alicenet/alicenet/layer1/chains/polygon"
)

var _ layer1.AllSmartContracts = &AllSmartContractsHandle{}

// AllSmartContractsHandle is a bus where we can access all smart contracts from the
// different layer1 clients.
type AllSmartContractsHandle struct {
	ethereumContracts *ethereum.Contracts
	polygonContracts  *polygon.Contracts
}

func NewAllSmartContractsHandle(
	ethClient layer1.Client,
	ethContractFactoryAddress common.Address,
	polygonClient layer1.Client,
	polygonContractFactoryAddress common.Address,
) layer1.AllSmartContracts {
	return &AllSmartContractsHandle{
		ethereumContracts: ethereum.NewContracts(ethClient, ethContractFactoryAddress),
		polygonContracts:  polygon.NewContracts(polygonClient, polygonContractFactoryAddress),
	}
}

func (ch *AllSmartContractsHandle) EthereumContracts() layer1.EthereumContracts {
	return ch.ethereumContracts
}

func (ch *AllSmartContractsHandle) PolygonContracts() layer1.MultichainContracts {
	return ch.polygonContracts
}
