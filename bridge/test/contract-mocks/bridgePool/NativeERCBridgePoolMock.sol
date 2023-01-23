// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/NativeERCBridgePoolBase.sol";
import "contracts/utils/auth/ImmutableSnapshots.sol";

contract NativeERCBridgePoolMock is Initializable, NativeERCBridgePoolBase, ImmutableSnapshots {
    address internal _ercContract;

    constructor(address bridgeRouterContract) NativeERCBridgePoolBase(bridgeRouterContract, _snapshotsAddress() ) {}

    function initialize(address ercContract_) public initializer {
        _ercContract = ercContract_;
    }

    function deposit(address sender, bytes calldata depositParameters_) public override {
        super.deposit(sender, depositParameters_);
    }

    function withdraw(
        address receiver,
        bytes memory vsPreImage,
        bytes[4] memory proofs
    ) public override returns (uint256 value) {
        value = super.withdraw(receiver, vsPreImage, proofs);
    }
}
