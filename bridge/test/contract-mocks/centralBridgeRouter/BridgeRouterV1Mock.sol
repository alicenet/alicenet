// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/ICentralBridgeRouter.sol";

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
        onlyCentralBridgeRouter
        returns (bytes memory)
    {
        msgSender_ = msgSender_;
        EventData[] memory eventData = new EventData[](1);
        DepositCallData memory receivedDepositData = abi.decode(data_, (DepositCallData));
        bytes32[] memory topics = new bytes32[](receivedDepositData.number);
        topics[0] = DepositedERCToken.selector;
        eventData[0] = EventData({topics: topics, logData: data_});
        DepositReturnData memory depositReturnData = DepositReturnData({
            eventData: eventData,
            fee: _fee
        });
        return abi.encode(depositReturnData);
    }
}
