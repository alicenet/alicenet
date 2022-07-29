// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library MagicTokenTransferErrors {
    error TransferFailed(address token, address to, uint256 amount);
}
