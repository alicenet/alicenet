// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/libraries/NFTStake/NFTStakeBase.sol";

/// @custom:salt StakeNFTLP
/// @custom:deploy-type deployUpgradeable
contract StakeNFTLP is NFTStakeBase {
    constructor() NFTStakeBase() {}

    function initialize() public onlyFactory initializer {
        __stakeNFTBaseInit("MNSNFT", "MNS");
    }
}
