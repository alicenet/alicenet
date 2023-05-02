// SPDX-License-Identifier: MIT
pragma solidity ^0.8.16;
import "contracts/utils/auth/ImmutableALCA.sol";
import "contracts/utils/auth/ImmutableFactory.sol";
import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "contracts/utils/ERC20SafeTransfer.sol";



abstract contract ALCAVaultBase is ImmutableALCA {
    IERC20 public immutable alca;
    error Erc20TransferFailed(address erc20Address, address from, address to, uint256 amount);
    error InsufficientBalance(uint256 balance, uint256 amount);
    constructor() ImmutableFactory(msg.sender){
        alca = IERC20(_alcaAddress());
    }
    function sendAlCA(address to_, uint256 amount_) public onlyFactory() {
        bool success = alca.transfer(to_, amount_);
        if (!success) {
            revert Erc20TransferFailed(
                _alcaAddress(),
                address(this),
                to_,
                amount_
            );
        }
    }

    function approveAllowance(address spender_, uint256 amount_) public onlyFactory() {
        uint256 balance = alca.balanceOf(address(this));
        if (amount_ > balance) {
            revert InsufficientBalance(balance, amount_);
        }
        alca.approve(spender_, amount_);
    }

    function getBalance() public view returns (uint256){
        return alca.balanceOf(address(this));
    }
}