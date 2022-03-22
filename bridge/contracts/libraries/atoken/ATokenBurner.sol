// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IAToken.sol";
import "contracts/utils/ImmutableAuth.sol";

/// @custom:salt ATokenBurner
/// @custom:deploy-type deployStatic
contract ATokenBurner is ImmutableAToken {
    // Placeholder contract. The real ATokenBurner will be implemented later
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() {}
}
