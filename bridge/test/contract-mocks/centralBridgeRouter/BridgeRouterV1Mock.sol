// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/ICentralBridgeRouter.sol";
import "hardhat/console.sol";

contract BridgeRouterV1Mock is ImmutableCentralBridgeRouter {
    struct EventData {
        bytes32[] topics;
        bytes logData;
    }

    struct DepositReturnData {
        EventData[] eventData;
        uint256 fee;
    }

    struct DepositCallData {
        address ercContract;
        uint8 destinationAccountType;
        address destinationAccount;
        uint8 tokenType;
        uint256 number;
        uint256 chainID;
        uint16 poolVersion;
    }

    uint256 internal immutable _fee;
    uint256 internal _dummy = 0;

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

    constructor(uint256 fee_) ImmutableFactory(msg.sender) {
        _fee = fee_;
    }

    function routeDeposit(address msgSender_, bytes memory data_)
        public
        view
        returns (
            //  onlyCentralBridgeRouter
            bytes memory
        )
    {
        uint256 depositNumber = abi.decode(data_, (uint256)); // we use depositNumber as size for the topics array
        msgSender_ = msgSender_;
        bytes32[] memory topics = new bytes32[](depositNumber);
        EventData[] memory eventData = new EventData[](1);
        topics[0] = DepositedERCToken.selector;
        eventData[0] = EventData({topics: topics, logData: data_});
        DepositReturnData memory depositReturnData = DepositReturnData({
            eventData: eventData,
            fee: _fee
        });
        return abi.encode(depositReturnData);
    }
}
