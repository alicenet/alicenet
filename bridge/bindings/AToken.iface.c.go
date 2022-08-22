// Generated by ifacemaker. DO NOT EDIT.

package bindings

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

// IATokenCaller ...
type IATokenCaller interface {
	// Allowance is a free data retrieval call binding the contract method 0xdd62ed3e.
	//
	// Solidity: function allowance(address owner, address spender) view returns(uint256)
	Allowance(opts *bind.CallOpts, owner common.Address, spender common.Address) (*big.Int, error)
	// BalanceOf is a free data retrieval call binding the contract method 0x70a08231.
	//
	// Solidity: function balanceOf(address account) view returns(uint256)
	BalanceOf(opts *bind.CallOpts, account common.Address) (*big.Int, error)
	// Convert is a free data retrieval call binding the contract method 0xa3908e1b.
	//
	// Solidity: function convert(uint256 amount) view returns(uint256)
	Convert(opts *bind.CallOpts, amount *big.Int) (*big.Int, error)
	// Decimals is a free data retrieval call binding the contract method 0x313ce567.
	//
	// Solidity: function decimals() view returns(uint8)
	Decimals(opts *bind.CallOpts) (uint8, error)
	// GetLegacyTokenAddress is a free data retrieval call binding the contract method 0x035c7099.
	//
	// Solidity: function getLegacyTokenAddress() view returns(address)
	GetLegacyTokenAddress(opts *bind.CallOpts) (common.Address, error)
	// GetMetamorphicContractAddress is a free data retrieval call binding the contract method 0x8653a465.
	//
	// Solidity: function getMetamorphicContractAddress(bytes32 _salt, address _factory) pure returns(address)
	GetMetamorphicContractAddress(opts *bind.CallOpts, _salt [32]byte, _factory common.Address) (common.Address, error)
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
