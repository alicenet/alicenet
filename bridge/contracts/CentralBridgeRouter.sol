// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePoolRouter.sol";
import "contracts/interfaces/ICentralBridgeRouter.sol";
import "contracts/libraries/errors/CentralBridgeRouterErrors.sol";

contract CentralBridgeRouter is ICentralBridgeRouter, ImmutableFactory, ImmutableBToken {
    
    struct EventData {
        bytes32[] topics;
        bytes logData;
    }
    
    struct DepositData {
        address PoolAddress;
        uint8 poolType;
        uint8 ercType;
        uint256 poolVersion;
        bytes depositDetails;
    }

    uint256 internal constant _TRIPPED_CP_VALUE = 2**256-1;
    
    // mapping of pool version to mapping of pool group to
    /*
    version native = 0, external = 1, lazymint = 2 erc20 = 0 erc721 = 1 erc1155 = 2

    00
    01
    02
    10
    11
    12

    1
    */ 
   /**
    * @dev mapping of pool type to fee  
    */
    mapping(uint32 => uint256) internal _poolTypeToFee;
    //TODO determine if we should have a default fee for resetting circuit breaker
    uint256 internal _nonce;
    //tracker to track number of deployed router versions
    uint16 internal _routerVersions;

    constructor() ImmutableFactory(msg.sender) ImmutableBToken() {}

    /**
     * takes token deposit calls from ALCB and emits deposit events on token transfer completion
     * @param msgSender_ address of account depositing tokens from
     * @param poolVersion_ version of pool to deposit token on
     * @param depositData_ abi encoded input data that describes the transaction
     */
    function depositNativeToken(
        address msgSender_,
        uint16 poolVersion_,
        address poolAddress,
        bytes calldata depositData_
    ) public onlyBToken returns (uint256 fee) {
        //get the router config for the version specified
        //get the router address
        DepositData memory depositData = abi.decode(depositData_, DepositData);
        fee = _getFee(depositData.poolVersion, depositData.poolType, depositData.ercType);
        
        if (routerConfig.notOnline)
            revert CentralBridgeRouterErrors.DisabledPoolVersion(
                poolVersion_,
                routerConfig.routerAddress
            );
        bytes memory returnDataBytes = IBridgePool(poolAddress).deposit(
            msgSender_,
            depositData_
        );
        DepositReturnData memory returnData = abi.decode(returnDataBytes, (DepositReturnData));
        uint256 nonce = _nonce;
        for (uint256 i = 0; i < returnData.eventData.length; i++) {
            returnData.eventData[i].logData = abi.encodePacked(
                returnData.eventData[i].logData,
                nonce
            );
            _emitDepositEvent(returnData.eventData[i].topics, returnData.eventData[i].logData);
            nonce += 1;
        }
        _nonce = nonce;
        fee = returnData.fee;
    }

    
    function addNewBridgeType(uint16 version_, uint8 poolType_, uint8 tokenType_, uint256 fee_) public onlyFactory {
        uint32 poolType = abi.encode(version_, poolType_, tokenType_);
        _poolTypeToFee[poolType] = fee_;
    }
    
    function tripCB(uint16 version_, uint8 poolType_, uint8 tokenType_) public onlyCBSetter {
        uint32 poolType = abi.encode(version_, poolType_, tokenType_);
        _poolTypeToFee[poolType] = _TRIPPED_CP_VALUE;
    }

    //TODO change fee_ reference dependent on how we want to handle default fees possibly change to 0 for migration
    function resetCB(uint16 version_, uint8 poolType_, uint8 tokenType_, uint256 fee_) public onlyCBResetter {
        uint32 poolType = abi.encode(version_, poolType_, tokenType_);
        _poolTypeToFee[poolType] = fee_;
    }
    function reverPoolOffline(uint16 version_, uint8 poolType_, uint8 tokenType_) public view returns (bool) {
        uint32 poolType = abi.encode(version_, poolType_, tokenType_);
        uint256 fee = _poolTypeToFee[poolType];

    }
    function _getFee(uint16 version_, uint8 poolType_, uint8 tokenType_) internal view returns (uint256) {
        uint32 poolType = abi.encode(version_, poolType_, tokenType_);
        uint256 fee = _poolTypeToFee[poolType];
        if (fee == _TRIPPED_CP_VALUE) revert CentralBridgeRouterErrors.CircuitBreakerTripped();
        return fee;
    }

    function _emitDepositEvent(bytes32[] memory topics_, bytes memory eventData_) internal {
        if (topics_.length == 1) _log1Event(topics_, eventData_);
        else if (topics_.length == 2) _log2Event(topics_, eventData_);
        else if (topics_.length == 3) _log3Event(topics_, eventData_);
        else if (topics_.length == 4) _log4Event(topics_, eventData_);
        else revert CentralBridgeRouterErrors.MissingEventSignature();
    }

    function _log1Event(bytes32[] memory topics_, bytes memory eventData_) internal {
        bytes32 topic0 = topics_[0];
        assembly {
            log1(add(eventData_, 0x20), mload(eventData_), topic0)
        }
    }

    function _log2Event(bytes32[] memory topics_, bytes memory eventData_) internal {
        bytes32 topic0 = topics_[0];
        bytes32 topic1 = topics_[1];
        assembly {
            log2(add(eventData_, 0x20), mload(eventData_), topic0, topic1)
        }
    }

    function _log3Event(bytes32[] memory topics_, bytes memory eventData_) internal {
        bytes32 topic0 = topics_[0];
        bytes32 topic1 = topics_[1];
        bytes32 topic2 = topics_[2];
        assembly {
            log3(add(eventData_, 0x20), mload(eventData_), topic0, topic1, topic2)
        }
    }

    function _log4Event(bytes32[] memory topics_, bytes memory eventData_) internal {
        bytes32 topic0 = topics_[0];
        bytes32 topic1 = topics_[1];
        bytes32 topic2 = topics_[2];
        bytes32 topic3 = topics_[3];
        assembly {
            log4(add(eventData_, 0x20), mload(eventData_), topic0, topic1, topic2, topic3)
        }
    }

    /*     function _isValidVersion(uint16 version_) internal view returns (bool) {
        RouterConfig memory config = _routerConfig[version_];
        if (config.routerAddress == address(0) || config.notOnline) return false;
        return true;
    } */
}
