// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/legacy/interfaces/IStakingToken.sol";
import "contracts/legacy/utils/auth/ImmutableFactory.sol";
import "contracts/legacy/utils/auth/ImmutableALCA.sol";

contract ALCABurnerMock is ImmutableALCA {
    constructor() ImmutableFactory(msg.sender) ImmutableALCA() {}

    function burn(address to, uint256 amount) public {
        IStakingToken(_alcaAddress()).externalBurn(to, amount);
    }
}
