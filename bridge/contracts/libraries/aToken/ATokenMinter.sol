// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IAToken.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts/access/AccessControl.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

import "hardhat/console.sol";

contract ATokenMinter is Initializable, ImmutableAToken {
    bytes32 public constant MINTER_ROLE = keccak256("MINTER_ROLE");
    address internal _admin;
    address internal _minter;

    modifier onlyAdmin() {
        require(msg.sender == _admin, "onlyAdmin");
        _;
    }

    modifier onlyMinter() {
        require(msg.sender == _minter, "onlyMinter");
        _;
    }

    constructor() ImmutableFactory(msg.sender) ImmutableAToken() {
        _admin = msg.sender;
    }

    function setAdmin(address admin_) public onlyFactory {
        _admin = admin_;
    }

    function setMinter(address minter_) public onlyAdmin {
        _minter = minter_;
    }

    function mint(address to, uint256 amount) public onlyMinter {
        // TODO check for this approach (classical Access Control)
        IAToken(_ATokenAddress()).externalMint(to, amount);
    }
}
