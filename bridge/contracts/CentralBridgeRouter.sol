// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePoolRouter.sol";
import "contracts/libraries/errors/CentralBridgeRouterErrors.sol";

contract CentralBridgeRouter is ImmutableFactory, ImmutableBToken {
    struct RouterConfig {
        address routerAddress;
        bool notOnline;
    }

    struct DepositReturnData {
        bytes32[] topics;
        bytes logData;
        uint256 fee;
    }
    uint16 internal _routerVersions;

    // mapping of router version to data
    mapping(uint16 => RouterConfig) internal _routerConfig;
    uint16 internal constant _POOL_VERSION = 1;
    uint256 internal immutable _chainID;
    uint256 internal _nonce;
    event DepositedERCToken(
        uint256 nonce,
        address ercContract,
        uint8 destinationAccountType, // 1 for secp256k1, 2 for bn128
        address destinationAccount, //account to deposit the tokens to in alicenet
        uint8 ercType,
        uint256 number, // If fungible, this is the amount. If non-fungible, this is the id
        uint256 chainID,
        uint16 poolVersion
    );

    constructor() ImmutableFactory(msg.sender) {
        _chainID = block.chainid;
    }

    function sendDeposit(
        address msgSender_,
        uint16 poolVersion_,
        bytes calldata depositData_
    ) public onlyBToken returns (uint256 fee) {
        //get the router config for the version specified
        //get the router address
        RouterConfig memory routerConfig = _routerConfig[poolVersion_];
        if (routerConfig.routerAddress == address(0))
            revert CentralBridgeRouterErrors.PoolVersionNotSupported(
                poolVersion_,
                routerConfig.routerAddress
            );
        if (routerConfig.notOnline)
            revert CentralBridgeRouterErrors.PoolVersionNotSupported(
                poolVersion_,
                routerConfig.routerAddress
            );
        bytes memory returnDataBytes = IBridgePoolRouter(routerConfig.routerAddress).routeDeposit(
            msgSender_,
            depositData_
        );
        DepositReturnData memory returnData = abi.decode(returnDataBytes, (DepositReturnData));
        _emitDepositEvent(returnData.topics, returnData.logData);
        fee = returnData.fee;
    }

    function addRouter(address newRouterAddress_) public onlyFactory {
        uint16 version = _routerVersions + 1;
        _routerConfig[version] = RouterConfig(newRouterAddress_, true);
        //update routerVersions tracker
        _routerVersions = version;
    }

    function _disableRouter(uint16 routerVersion_) internal {}

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
