// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library ETHSafeTransferErrors {
    error CannotTransferToZeroAddress();
    error EthTransferFailed(address from, address to, uint256 amount);
}
