// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IStakingToken.sol";
import "contracts/utils/ImmutableAuth.sol";

contract ATokenBurnerMock is ImmutableAToken {
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() {}

    function burn(address to, uint256 amount) public {
        IStakingToken(_aTokenAddress()).externalBurn(to, amount);
    }
}
