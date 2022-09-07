// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/utils/ImmutableAuth.sol";
import "contracts/LocalERC20BridgePoolV1.sol";
import "contracts/LocalERC721BridgePoolV1.sol";
import "contracts/LocalERC1155BridgePoolV1.sol";

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "hardhat/console.sol";

/// @custom:salt BridgeRouter
/// @custom:deploy-type deployStatic
contract BridgeRouter is
    Initializable,
    ImmutableFactory,
    ImmutableLocalERC20BridgePoolV1,
    ImmutableLocalERC721BridgePoolV1,
    ImmutableLocalERC1155BridgePoolV1
{
    uint256 internal immutable _fee;
    uint256 internal _dummy = 0;
    uint256 internal _chainId = 1337;

    struct DepositCallData {
        address ercContract;
        uint8 destinationAccountType;
        address destinationAccount;
        uint8 tokenType;
        uint256 number;
        uint256 chainID;
        uint16 poolVersion;
    }

    constructor(uint256 fee_) ImmutableFactory(msg.sender) {
        _fee = fee_;
    }

    function initialize() public onlyFactory initializer {}

    function routeDeposit(address account, bytes calldata data) external returns (uint256) {
        DepositCallData memory depositCallData = abi.decode(data, (DepositCallData));
        if (depositCallData.tokenType == 1) {
            LocalERC20BridgePoolV1(_localERC20BridgePoolV1Address()).deposit(
                account,
                depositCallData.number
            );
        } else if (depositCallData.tokenType == 2) {
            LocalERC721BridgePoolV1(_localERC721BridgePoolV1Address()).deposit(
                account,
                depositCallData.number
            );
        } else if (depositCallData.tokenType == 3) {
            LocalERC1155BridgePoolV1(_localERC1155BridgePoolV1Address()).deposit(
                account,
                depositCallData.number,
                1
            );
        }
        account = account;
        data = data;
        _dummy = 0;
        return _fee;
    }
}
