// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

contract BridgeRouterMock {
    struct DepositReturnData {
        bytes32[] topics;
        bytes logData;
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

    uint256 internal immutable _fee;
    uint256 internal _dummy = 0;

    constructor(uint256 fee_) {
        _fee = fee_;
    }

    function routeDeposit(address msgSender_, bytes memory data_) public onlyCentralRouter returns (bytes memory) {
        msgSender_= msgSender_;
        bytes32[] memory topics;
        topics[0] = DepositedERCToken.selector;
        DepositReturnData memory depositReturnData = DepositReturnData({
            topics: topics,
            logData: data_,
            fee: _fee
        });
        return abi.encode(depositReturnData);
    }
}