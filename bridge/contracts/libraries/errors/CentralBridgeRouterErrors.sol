// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

library CentralBridgeRouterErrors {
    error PoolVersionNotSupported(uint16 version, address routerAddress);
    error MissingEventSignature();
}
