// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

interface IBridgePoolRouter {

    function routeDeposit(address account, bytes calldata data) external returns (bytes memory);

}
