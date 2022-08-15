// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

// Struct to keep track of the amount divided between the staking contracts
// dividends distribution.
struct Splits {
    uint32 validatorStaking;
    uint32 publicStaking;
    uint32 liquidityProviderStaking;
    uint32 protocolFee;
}

interface IDistribution {}
