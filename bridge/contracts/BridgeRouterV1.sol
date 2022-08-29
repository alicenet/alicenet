// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/AliceNetFactory.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/libraries/errorCodes/BridgeRouterErrorCodes.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errorCodes/CircuitBreakerErrorCodes.sol";
import "hardhat/console.sol";

/// @custom:salt BridgeRouter
/// @custom:deploy-type deployUpgradeable
contract BridgeRouter is
    Initializable,
    ImmutableFactory,
    ImmutableBridgeRouter,
    ImmutableBridgePoolDepositNotifier,
    ImmutableBToken
{
    event DepositedERCToken(
        uint256 nonce,
        address ercContract,
        address owner,
        uint8 ERCType,
        uint256 number, // If fungible, this is the amount. If non-fungible, this is the id
        uint256 chainID
    );

    struct DepositCallData {
        address ERCContract;
        uint8 tokenType;
        uint256 number;
        uint256 chainID;
        uint16 poolVersion;
    }
    uint16 constant POOL_VERSION = 1;
    uint256 internal immutable _networkId;
    
    uint256 nonce;
    bool private publicPoolDeploymentEnabled = false;

    constructor(uint256 networkId_) ImmutableFactory(msg.sender) {
        _networkId = networkId_;
    }

    /**
     * @dev this function can only be called by the btoken contract, when called this function will
     * calculate the target pool address with the information in call data.
     * The fee will be passed in as a parameter by the user and btoken verify the deposit happened before
     * forwarding the call to this function.
     * @param msgSender the original sender of the value
     * @param maxTokens max number of bTokens willing to pay for the deposit
     * @param data contains information needed to perform deposit, ERCCONTRACTADDRESS, ChainID, Version
     */
    function routeDeposit(
        address msgSender,
        uint256 maxTokens,
        bytes memory data
    ) public onlyBToken returns (uint256 btokenFeeAmount) {
        //get the fee to deposit a token into the bridge
        btokenFeeAmount = 1000; // TODO: @gus get proper value for bToken fee
        require(
            maxTokens >= btokenFeeAmount,
            string(abi.encodePacked(BridgeRouterErrorCodes.BRIDGEROUTER_INSUFFICIENT_FUNDS))
        );
        // use abi decode to extract the information out of data
        DepositCallData memory depositCallData = abi.decode(data, (DepositCallData));
        //encode the salt with the information from
        bytes32 poolSalt = getBridgePoolSalt(
            depositCallData.ERCContract,
            depositCallData.tokenType,
            depositCallData.chainID,
            POOL_VERSION
        );
        // calculate the address of the pool
        address poolAddress = getStaticPoolContractAddress(poolSalt, address(this));

        //call the pool to initiate deposit
        IBridgePool(poolAddress).deposit(msgSender, depositCallData.number);

        emit DepositedERCToken(
            nonce,
            depositCallData.ERCContract,
            msg.sender,
            depositCallData.tokenType,
            depositCallData.number,
            _networkId
        );
        nonce++;
        return btokenFeeAmount;
    }
}
