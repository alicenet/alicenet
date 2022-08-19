// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

interface IBridgeRouter {
    function routeDeposit(address account, bytes calldata data) external returns (uint256);
}
