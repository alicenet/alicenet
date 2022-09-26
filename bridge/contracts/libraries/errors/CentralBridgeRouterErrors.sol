// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

library CentralBridgeRouterErrors {
    error InvalidPoolVersion(uint16 version);
    error DisabledPoolVersion(uint16 version, address routerAddress);
    error MissingEventSignature();
}
