// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/AliceNetFactory.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/libraries/errorCodes/BridgeRouterErrorCodes.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/utils/BridgePoolAddressUtil.sol";
import "contracts/libraries/BridgeRouterV1Errors.sol";

/// @custom:salt BridgePoolRouterV1
/// @custom:deploy-type deployStatic
contract BridgePoolRouterV1 is
    Initializable,
    ImmutableFactory,
    ImmutableBridgeRouter,
    ImmutableBridgePoolDepositNotifier,
    ImmutableBToken
{
    event DepositedERCToken(
        uint256 nonce,
        address ercContract,
        uint8 destinationAccountType, // 1 for secp256k1, 2 for bn128
        address destinationAccount, //account to deposit the tokens to in alicenet
        uint8 ERCType,
        uint256 number, // If fungible, this is the amount. If non-fungible, this is the id
        uint256 chainID,
        uint16 poolVersion
    );
    struct Config {
        bool depositPaused;
        uint64 depositFee;
    }
    struct DepositCallData {
        address ERCContract;
        uint8 destinationAccountType;
        address destinationAccount;
        uint8 tokenType;
        uint256 number;
        uint256 chainID;
        uint16 poolVersion;
    }
    uint16 constant POOL_VERSION = 1;
    uint256 internal immutable _networkId;
    uint256 nonce;
    Config internal _configs;

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
        bytes memory data
    ) public onlyBToken returns (uint256 btokenFeeAmount) {
        Config memory configs = _configs;
        if(config.depositPaused){
            revert(BridgeRouterV1Errors.BridgeDepositsPaused());
        }
        // use abi decode to extract the information out of data
        DepositCallData memory depositCallData = abi.decode(data, (DepositCallData));
        //encode the salt with the information from
        bytes32 poolSalt = BridgePoolAddressUtil.getBridgePoolSalt(
            depositCallData.ERCContract,
            depositCallData.tokenType,
            depositCallData.chainID,
            POOL_VERSION
        );
        // calculate the address of the pool
        address poolAddress = BridgePoolAddressUtil.getBridgePoolAddress(poolSalt, _bridgeFactoryAddress());
        //call the pool to initiate deposit
        IBridgePool(poolAddress).deposit(msgSender, depositCallData.number);
        //TODO determine if we are tracking current networkID or token NetworkID
        emit DepositedERCToken(
            nonce,
            depositCallData.ERCContract,
            depositCallData.destinationAccountType,
            depositCallData.destinationAccount,
            depositCallData.tokenType,
            depositCallData.number,
            _networkId
        );
        nonce++;
        return configs.depositFee;
    }

    /**
     * function to set BtokenFee required for deposit
     */
    function setDepositFee(uint64 newDepositFee_) public onlyFactory{
        config.depositFee = newDepositFee_;
    }
    /**
     * mechanism to allow alicenet factory to pause all deposits going into the pool 
     * 
     */
    function pauseAllDeposits()public onlyFactory{
        //TODO determine if i should do a check if this is already set to true
        //I dont really see a point in burning gas to check if you are already paused 
        config.depositPaused = true;
    }

    /**
     * allows the factory to allow deposits on to bridge pools
     */
    function resumeAllDeposits() public onlyFactory{
        config.depositPaused = false;
    }

}
