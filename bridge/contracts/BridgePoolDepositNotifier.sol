// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "contracts/utils/ImmutableAuth.sol";
import {BridgeRouter} from "contracts/BridgeRouterV1.sol";

/// @custom:salt BridgePoolDepositNotifier
/// @custom:deploy-type deployUpgradeable
contract BridgePoolDepositNotifier is ImmutableFactory, ImmutableBridgeRouter {
    uint256 internal _nonce = 0;
    uint256 internal immutable _networkId;

    event Deposited(
        uint256 nonce,
        address ercContract,
        address owner,
        uint8 tokenType,
        uint256 number, // If fungible, this is the amount. If non-fungible, this is the id
        uint256 networkId
    );

    /**
     * @notice onlyBridgePool verifies that the call is done by one BridgePools intanciated by BridgeRouter
     * @param bridgePoolSalt informed salt
     */
    modifier onlyBridgePool(bytes32 bridgePoolSalt) {
        address allowedAddress = BridgeRouter(_bridgeRouterAddress()).getStaticPoolContractAddress(
            bridgePoolSalt,
            _bridgeRouterAddress()
        );
        require(
            msg.sender == allowedAddress,
            string(abi.encodePacked(ImmutableAuthErrorCodes.IMMUTEABLEAUTH_ONLY_BRIDGEPOOL))
        );
        _;
    }

    constructor(uint256 networkId_) ImmutableFactory(msg.sender) {
        _networkId = networkId_;
    }

    /**
     * @notice doEmit emit a deposit event with params informed
     * @param salt calculated salt of the caller contract
     * @param ercContract ERC contract
     * @param tokenType 1=ERC20, 2=ERC721
     * @param owner address for deposit
     * @param number amount of tokens or tokenId
     */
    function doEmit(
        bytes32 salt,
        address ercContract,
        address owner,
        uint8 tokenType,
        uint256 number
    ) public onlyBridgePool(salt) {
        uint256 n = _nonce + 1;
        emit Deposited(n, ercContract, owner, tokenType, number, _networkId);
        _nonce = n;
    }
}
