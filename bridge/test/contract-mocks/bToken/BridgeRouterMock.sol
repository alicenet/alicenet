// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/utils/ImmutableAuth.sol";
import "contracts/LocalERC20BridgePoolV1.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "hardhat/console.sol";


/// @custom:salt BridgeRouter
/// @custom:deploy-type deployStatic
contract BridgeRouter is     Initializable,
 ImmutableFactory, ImmutableLocalERC20BridgePoolV1   {
    uint256 internal immutable _fee;
    uint256 internal _dummy = 0;
    uint256 internal _chainId = 1337;

    constructor(uint256 fee_) ImmutableFactory(msg.sender) {
        _fee = fee_;
    }

    function initialize() public onlyFactory initializer {}

    function routeDeposit(address account, bytes calldata data) external returns (uint256) {
        uint256 amount = 100;
         LocalERC20BridgePoolV1(_localERC20BridgePoolV1Address()).deposit(account, amount);
         account = account;
        data = data;
        _dummy = 0;
        return _fee;
    }
}
