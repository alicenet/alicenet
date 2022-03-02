// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "@openzeppelin/contracts-upgradeable/token/ERC721/ERC721Upgradeable.sol";
import "@openzeppelin/contracts-upgradeable/token/ERC721/IERC721Upgradeable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/utils/EthSafeTransfer.sol";
import "contracts/utils/ERC20SafeTransfer.sol";
import "contracts/utils/MagicValue.sol";

/// @custom:salt Foundation
/// @custom:deploy-type deployUpgradeable
contract Foundation is
    Initializable,
    MagicValue,
    EthSafeTransfer,
    ERC20SafeTransfer,
    ImmutableFactory,
    ImmutableMadToken
{
    constructor() ImmutableFactory(msg.sender) ImmutableMadToken() {}

    function initialize() public initializer onlyFactory {}

    /// DO NOT CALL THIS METHOD UNLESS YOU ARE MAKING A DISTRIBUTION AS ALL VALUE
    /// WILL BE DISTRIBUTED TO STAKERS EVENLY. depositToken distributes MadToken
    /// to all stakers evenly should only be called during a slashing event. Any
    /// MadToken sent to this method in error will be lost. This function will
    /// fail if the circuit breaker is tripped. The magic_ parameter is intended
    /// to stop some one from successfully interacting with this method without
    /// first reading the source code and hopefully this comment
    function depositToken(uint8 magic_, uint256 amount_) public checkMagic(magic_) {
        // collect tokens
        _safeTransferFromERC20(IERC20Transferable(_MadTokenAddress()), msg.sender, amount_);
    }

    /// DO NOT CALL THIS METHOD UNLESS YOU ARE MAKING A DISTRIBUTION ALL VALUE
    /// WILL BE DISTRIBUTED TO STAKERS EVENLY depositEth distributes Eth to all
    /// stakers evenly should only be called by MadBytes contract any Eth sent to
    /// this method in error will be lost this function will fail if the circuit
    /// breaker is tripped the magic_ parameter is intended to stop some one from
    /// successfully interacting with this method without first reading the
    /// source code and hopefully this comment
    function depositEth(uint8 magic_) public payable checkMagic(magic_) {}
}
