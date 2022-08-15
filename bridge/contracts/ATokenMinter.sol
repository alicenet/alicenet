// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IStakingToken.sol";
import "contracts/utils/ImmutableAuth.sol";

/// @custom:salt ATokenMinter
/// @custom:deploy-type deployUpgradeable
contract ATokenMinter is ImmutableAToken, IStakingTokenMinter {
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() IStakingTokenMinter() {}

    function mint(address to, uint256 amount) public onlyFactory {
        IStakingToken(_aTokenAddress()).externalMint(to, amount);
    }
}
