// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IAToken.sol";
import "contracts/utils/ImmutableAuth.sol";

/// @custom:salt ATokenMinter
/// @custom:deploy-type onlyProxy
contract ATokenMinter is ImmutableAToken {
    // Placeholder contract. The real ATokenMinter will be implemented later
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() {}
}
