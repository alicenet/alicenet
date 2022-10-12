// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePoolRouter.sol";
import "contracts/interfaces/ICentralBridgeRouter.sol";
import "contracts/libraries/errors/CentralBridgeRouterErrors.sol";

contract CentralBridgeRouter is ICentralBridgeRouter, ImmutableFactory, ImmutableBToken {
    // mapping of router version to data
    mapping(uint16 => RouterConfig) internal _routerConfig;
    uint256 internal _nonce;
    //tracker to track number of deployed router versions
    uint16 internal _routerVersions;

    event DepositedERCToken(
        address ercContract,
        uint8 destinationAccountType, // 1 for secp256k1, 2 for bn128
        address destinationAccount, //account to deposit the tokens to in alicenet
        uint8 ercType,
        uint256 number, // If fungible, this is the amount. If non-fungible, this is the id
        uint256 chainID,
        uint16 poolVersion,
        uint256 nonce
    );

    constructor() ImmutableFactory(msg.sender) ImmutableBToken() {}

    /**
     * takes token deposit calls from ALCB and emits deposit events on token transfer completion
     * @param msgSender_ address of account depositing tokens from
     * @param poolVersion_ version of pool to deposit token on
     * @param depositData_ abi encoded input data that describes the transaction
     */
    function forwardDeposit(
        address msgSender_,
        uint16 poolVersion_,
        bytes calldata depositData_
    ) public onlyBToken returns (uint256 fee) {
        //get the router config for the version specified
        //get the router address
        RouterConfig memory routerConfig = _routerConfig[poolVersion_];
        if (routerConfig.routerAddress == address(0))
            revert CentralBridgeRouterErrors.InvalidPoolVersion(poolVersion_);
        if (routerConfig.notOnline)
            revert CentralBridgeRouterErrors.DisabledPoolVersion(
                poolVersion_,
                routerConfig.routerAddress
            );
        bytes memory returnDataBytes = IBridgePoolRouter(routerConfig.routerAddress).routeDeposit(
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

    /**
     * adds a new router to the routerConfig mapping
     * @param newRouterAddress_ address of the new router being added
     */
    function addRouter(address newRouterAddress_) public onlyFactory {
        uint16 version = _routerVersions + 1;
        _routerConfig[version] = RouterConfig(newRouterAddress_, false);
        //update routerVersions tracker
        _routerVersions = version;
    }

    /**
     * allows factory to disable deposits to routers
     * @param routerVersion_ version of router to disable
     */
    function disableRouter(uint16 routerVersion_) public onlyFactory {
        RouterConfig memory config = _routerConfig[routerVersion_];
        if (config.routerAddress == address(0))
            revert CentralBridgeRouterErrors.InvalidPoolVersion(routerVersion_);
        if (config.notOnline) {
            revert CentralBridgeRouterErrors.DisabledPoolVersion(
                routerVersion_,
                config.routerAddress
            );
        }
        _routerConfig[routerVersion_].notOnline = true;
    }

    /**
     * getter function for retrieving number of existing routers
     */
    function getRouterCount() public view returns (uint16) {
        return _routerVersions;
    }

    /**
     * getter function for getting address of online versioned router
     * @param routerVersion_ version of router to query
     * @return routerAddress address of versioned router
     */
    function getRouterAddress(uint16 routerVersion_) public view returns (address routerAddress) {
        RouterConfig memory config = _routerConfig[routerVersion_];
        if (config.routerAddress == address(0))
            revert CentralBridgeRouterErrors.InvalidPoolVersion(routerVersion_);
        routerAddress = config.routerAddress;
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
