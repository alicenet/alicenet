// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/interfaces/IStakingToken.sol";
import "contracts/utils/auth/ImmutableFactory.sol";
import "contracts/utils/auth/ImmutableAToken.sol";

contract ATokenMinterMock is ImmutableAToken {
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() {}

    function mint(address to, uint256 amount) public {
        IStakingToken(_aTokenAddress()).externalMint(to, amount);
    }
}
