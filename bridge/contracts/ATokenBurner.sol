// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/interfaces/IStakingToken.sol";
import "contracts/utils/auth/ImmutableFactory.sol";
import "contracts/utils/auth/ImmutableAToken.sol";

/// @custom:salt ATokenBurner
/// @custom:deploy-type deployUpgradeable
contract ATokenBurner is ImmutableAToken, IStakingTokenBurner {
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() IStakingTokenBurner() {}

    /**
     * @notice Burns ATokens
     * @param from_ The address from where the tokens will be burned
     * @param amount_ The amount of ATokens to be burned
     */
    function burn(address from_, uint256 amount_) public onlyFactory {
        IStakingToken(_aTokenAddress()).externalBurn(from_, amount_);
    }
}
