// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IStakingToken.sol";
import "contracts/utils/ImmutableAuth.sol";

contract ATokenMinterMock is ImmutableAToken {
    constructor() ImmutableFactory(msg.sender) ImmutableAToken() {}

    function mint(address to, uint256 amount) public {
        IStakingToken(_aTokenAddress()).externalMint(to, amount);
    }
}
