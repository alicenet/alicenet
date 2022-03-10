// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IAToken.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract ATokenBurner is ImmutableAToken, AccessControl {
    bytes32 public constant BURNER_ROLE = keccak256("BURNER_ROLE");

    constructor() ImmutableFactory(msg.sender) ImmutableAToken() {}

    function burn(address to, uint256 amount) public onlyRole(BURNER_ROLE) {
        IAToken(_ATokenAddress()).externalBurn(to, amount);
    }
}
