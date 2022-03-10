// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IAToken.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";

contract ATokenBurner is ImmutableAToken {
    address internal _admin;
    address internal _burner;

    modifier onlyAdmin() {
        require(msg.sender == _admin, "onlyAdmin");
        _;
    }

    modifier onlyBurner() {
        require(msg.sender == _burner, "onlyBurner");
        _;
    }

    constructor() ImmutableFactory(msg.sender) ImmutableAToken() {
        _admin = msg.sender;
    }

    function setAdmin(address admin_) public onlyFactory {
        _admin = admin_;
    }

    function setBurner(address burner_) public onlyAdmin {
        _burner = burner_;
    }

    function burn(address to, uint256 amount) public onlyBurner {
        IAToken(_ATokenAddress()).externalBurn(to, amount);
    }
}
