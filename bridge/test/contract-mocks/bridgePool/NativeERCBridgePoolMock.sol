// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/NativeERCBridgePoolBase.sol";

contract NativeERCBridgePoolMock is Initializable, NativeERCBridgePoolBase {
    address internal _ercContract;

    function initialize(address ercContract_) public onlyFactory initializer {
        _ercContract = ercContract_;
    }

    function deposit(address msgSender, bytes calldata depositParameters_) public override {
        super.deposit(msgSender, depositParameters_);
    }

    function withdraw(bytes memory vsPreImage, bytes[4] memory proofs)
        public
        override
        returns (address account, uint256 value)
    {
        (address account, uint256 value) = super.withdraw(vsPreImage, proofs);
        return (account, value);
    }
}
