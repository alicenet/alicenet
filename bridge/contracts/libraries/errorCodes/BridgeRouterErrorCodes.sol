// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library BridgeRouterErrorCodes {
    bytes32 public constant BRIDGEROUTER_UNEXISTENT_BRIDGEPOOL_IMPLEMENTATION_VERSION = "2500";
    bytes32 public constant BRIDGEROUTER_UNABLE_TO_DEPLOY_BRIDGEPOOL = "2501";
    bytes32 public constant BRIDGEROUTER_INSUFFICIENT_FUNDS = "2502";
}
