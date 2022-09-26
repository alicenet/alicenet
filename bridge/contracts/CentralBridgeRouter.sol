// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePoolRouter.sol";
import "contracts/interfaces/ICentralBridgeRouter.sol";
import "contracts/libraries/errors/CentralBridgeRouterErrors.sol";


contract CentralBridgeRouter is ICentralBridgeRouter ,ImmutableFactory, ImmutableBToken {
    struct RouterConfig {
        address routerAddress;
        bool notOnline;
    }

    // mapping of router version to data
    mapping(uint16 => RouterConfig) internal _routerConfig;
    uint256 internal _nonce;
    //tracker to track number of deployed router versions
    uint16 internal _routerVersions;

    constructor() ImmutableFactory(msg.sender) {}

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
        DepositReturnData memory returnData = abi.decode(
            returnDataBytes,
            (DepositReturnData)
        );
        uint256 nonce = _nonce;
        for(uint256 i =0; i < returnData.eventData.length; i++){
            returnData.eventData[i].logData = abi.encode(returnData.eventData[i].logData, nonce);
            _emitDepositEvent(returnData.eventData[i].topics, returnData.eventData[i].logData);
            nonce += 1;
        }
        _nonce = nonce;
        fee = returnData.fee;
    }

    function addRouter(address newRouterAddress_) public onlyFactory {
        uint16 version = _routerVersions + 1;
        _routerConfig[version] = RouterConfig(newRouterAddress_, false);
        //update routerVersions tracker
        _routerVersions = version;
    }

    function disableRouter(uint16 routerVersion_) public onlyFactory {
        RouterConfig memory config = _routerConfig[routerVersion_];
        if (config.notOnline) {
            revert CentralBridgeRouterErrors.DisabledPoolVersion(
                routerVersion_,
                config.routerAddress
            );
        }
        _routerConfig[routerVersion_].notOnline = true;
    }

    function getRouterCount() public view returns (uint16) {
        return _routerVersions;
    }

    function getRouterAddress(uint16 routerVersion_) public view returns (address routerAddress) {
        RouterConfig memory config = _routerConfig[routerVersion_];
        if (config.routerAddress == address(0)) revert CentralBridgeRouterErrors.InvalidPoolVersion(routerVersion_);
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

    function _isValidVersion(uint16 version_) internal view returns (bool) {
        RouterConfig memory config = _routerConfig[version_];
        if (config.routerAddress == address(0) || config.notOnline) return false;
        return true;
    }
}
