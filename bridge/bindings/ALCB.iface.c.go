// Generated by ifacemaker. DO NOT EDIT.

package bindings

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// IALCBCaller ...
type IALCBCaller interface {
	// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
	//
	// Solidity: function allowance(address owner, address spender) view returns(uint256)
	Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error)
	// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
	//
	// Solidity: function balanceOf(address account) view returns(uint256)
	BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error)
	// Decimals is a free data retrieval call binding the contract method 0x313ce567.
	//
	// Solidity: function decimals() view returns(uint8)
	Decimals(opts *bind.CallOpts) (uint8, error)
	// GetCentralBridgeRouterAddress is a free data retrieval call binding the contract method 0xff32fefc.
	//
	// Solidity: function getCentralBridgeRouterAddress() view returns(address)
	GetCentralBridgeRouterAddress(opts *bind.CallOpts) (common.Address, error)
	// GetDeposit is a free data retrieval call binding the contract method 0x9f9fb968.
	//
	// Solidity: function getDeposit(uint256 depositID) view returns((uint8,address,uint256))
	GetDeposit(opts *bind.CallOpts, depositID *big.Int) (Deposit, error)
	// GetDepositID is a free data retrieval call binding the contract method 0x94f70c8a.
	//
	// Solidity: function getDepositID() view returns(uint256)
	GetDepositID(opts *bind.CallOpts) (*big.Int, error)
	// GetEthFromALCBsBurn is a free data retrieval call binding the contract method 0x0b6774a1.
	//
	// Solidity: function getEthFromALCBsBurn(uint256 poolBalance_, uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
	GetEthFromALCBsBurn(opts *bind.CallOpts, poolBalance_ *big.Int, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error)
	// GetEthToMintALCBs is a free data retrieval call binding the contract method 0x0619c2f3.
	//
	// Solidity: function getEthToMintALCBs(uint256 totalSupply_, uint256 numBTK_) pure returns(uint256 numEth)
	GetEthToMintALCBs(opts *bind.CallOpts, totalSupply_ *big.Int, numBTK_ *big.Int) (*big.Int, error)
	// GetLatestEthFromALCBsBurn is a free data retrieval call binding the contract method 0xff326a2c.
	//
	// Solidity: function getLatestEthFromALCBsBurn(uint256 numBTK_) view returns(uint256 numEth)
	GetLatestEthFromALCBsBurn(opts *bind.CallOpts, numBTK_ *big.Int) (*big.Int, error)
	// GetLatestEthToMintALCBs is a free data retrieval call binding the contract method 0x777acd84.
	//
	// Solidity: function getLatestEthToMintALCBs(uint256 numBTK_) view returns(uint256 numEth)
	GetLatestEthToMintALCBs(opts *bind.CallOpts, numBTK_ *big.Int) (*big.Int, error)
	// GetLatestMintedALCBsFromEth is a free data retrieval call binding the contract method 0x71497ca6.
	//
	// Solidity: function getLatestMintedALCBsFromEth(uint256 numEth_) view returns(uint256)
	GetLatestMintedALCBsFromEth(opts *bind.CallOpts, numEth_ *big.Int) (*big.Int, error)
	// GetMarketSpread is a free data retrieval call binding the contract method 0x086cfefd.
	//
	// Solidity: function getMarketSpread() pure returns(uint256)
	GetMarketSpread(opts *bind.CallOpts) (*big.Int, error)
	// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
	//
	// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
	GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error)
	// GetMintedALCBsFromEth is a free data retrieval call binding the contract method 0x6e1d3f22.
	//
	// Solidity: function getMintedALCBsFromEth(uint256 poolBalance_, uint256 numEth_) pure returns(uint256)
	GetMintedALCBsFromEth(opts *bind.CallOpts, poolBalance_ *big.Int, numEth_ *big.Int) (*big.Int, error)
	// GetPoolBalance is a free data retrieval call binding the contract method 0xabd70aa2.
	//
	// Solidity: function getPoolBalance() view returns(uint256)
	GetPoolBalance(opts *bind.CallOpts) (*big.Int, error)
	// GetTotalALCBsDeposited is a free data retrieval call binding the contract method 0x90813858.
	//
	// Solidity: function getTotalALCBsDeposited() view returns(uint256)
	GetTotalALCBsDeposited(opts *bind.CallOpts) (*big.Int, error)
	// GetYield is a free data retrieval call binding the contract method 0x7c262871.
	//
	// Solidity: function getYield() view returns(uint256)
	GetYield(opts *bind.CallOpts) (*big.Int, error)
	// Name is a free data retrieval call binding the contract method 0x06fdde03.
	//
	// Solidity: function name() view returns(string)
	Name(opts *bind.CallOpts) (string, error)
	// Symbol is a free data retrieval call binding the contract method 0x95d89b41.
	//
	// Solidity: function symbol() view returns(string)
	Symbol(opts *bind.CallOpts) (string, error)
	// TotalSupply is a free data retrieval call binding the contract method 0x18160ddd.
	//
	// Solidity: function totalSupply() view returns(uint256)
	TotalSupply(opts *bind.CallOpts) (*big.Int, error)
}