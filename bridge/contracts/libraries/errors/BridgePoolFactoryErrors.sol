// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
library BridgePoolFactoryErrors {
    error PublicPoolDeploymentDisabled();
    error PoolVersionNotSupported(uint16 version);
    error StaticPoolDeploymentFailed(bytes32 salt_);
}