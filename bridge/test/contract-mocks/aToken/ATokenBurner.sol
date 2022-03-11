// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IAToken.sol";
import "contracts/utils/ImmutableAuth.sol";

contract ATokenBurner is ImmutableAToken {
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() {}

    function burn(address to, uint256 amount) public {
        IAToken(_ATokenAddress()).externalBurn(to, amount);
    }
}
