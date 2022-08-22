// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

interface IMagicTokenTransfer {
    function depositToken(uint8 magic_, uint256 amount_) external payable;
}
