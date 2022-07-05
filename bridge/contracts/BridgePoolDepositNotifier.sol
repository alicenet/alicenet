// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/BridgePoolFactory.sol";

/// @custom:salt BridgePoolDepositNotifier
/// @custom:deploy-type deployUpgradeable
contract BridgePoolDepositNotifier is ImmutableFactory, ImmutableBridgePoolFactory {
    uint256 internal _nonce = 0;
    uint256 internal immutable _networkId;

    event Deposited(
        uint256 nonce,
        address ercContract,
        address owner,
        uint256 number, // If fungible, this is the amount. If non-fungible, this is the id
        uint256 networkId
    );

    /**
     * @notice onlyBridgePool verifies that the call is done by one BridgePools intanciated by BridgePoolFactory
     * @param informedSalt informed salt
     * @param erc20Contract_ ERC contract address
     */
    modifier onlyBridgePool(bytes32 informedSalt, address erc20Contract_) {
        // msg.sender == calculatedAddress
        bytes32 calculatedSalt = BridgePoolFactory(_bridgePoolFactoryAddress())
            .getSaltFromERC20Address(erc20Contract_);
        require(
            informedSalt == calculatedSalt,
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
     * @param number amount of tokens
     * @param owner msg.sender of deposit
     */
    function doEmit(
        bytes32 salt,
        address ercContract,
        uint256 number,
        address owner
    ) public onlyBridgePool(salt, ercContract) {
        uint256 n = _nonce + 1;
        emit Deposited(n, ercContract, owner, number, _networkId);
        _nonce = n;
    }
}
