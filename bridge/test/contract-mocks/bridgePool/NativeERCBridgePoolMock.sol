// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/NativeERCBridgePoolBase.sol";

contract NativeERCBridgePoolMock is
    Initializable,
    NativeERCBridgePoolBase,
    ImmutableBridgePoolFactory
{
    address internal _ercContract;

    constructor(address alicenetFactoryAddress, address snapshotsAddress)
        NativeERCBridgePoolBase(alicenetFactoryAddress, snapshotsAddress)
    {}

    function initialize(address ercContract_) public onlyBridgePoolFactory initializer {
        _ercContract = ercContract_;
    }

    function deposit(address msgSender, bytes calldata depositParameters_) public override {
        super.deposit(msgSender, depositParameters_);
    }

    function withdraw(
        address msgSender,
        bytes memory vsPreImage,
        bytes[4] memory proofs
    ) public override returns (address account, uint256 value) {
        (account, value) = super.withdraw(msgSender, vsPreImage, proofs);
    }
}
