// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

contract BridgeRouterMock {
    uint256 internal _dummy = 0;

    function routeDeposit(address account, bytes calldata data) external returns (uint256) {
        account = account;
        data = data;
        _dummy = 0;
        return 1000;
    }
}
