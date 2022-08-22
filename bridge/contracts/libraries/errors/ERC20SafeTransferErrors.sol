// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library ERC20SafeTransferErrors {
    error CannotCallContractMethodsOnZeroAddress();
    error Erc20TransferFailed(address erc20Address, address from, address to, uint256 amount);
}
