// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/LocalERCBridgePoolBase.sol";

contract LocalERCBridgePoolMock is Initializable, LocalERCBridgePoolBase {
    address internal _ercContract;

    function initialize(address ercContract_) public onlyFactory initializer {
        _ercContract = ercContract_;
    }

    function deposit(address msgSender, bytes calldata depositParameters_) public override {
        super.deposit(msgSender, depositParameters_);
    }

    function withdraw(bytes memory _txInPreImage, bytes[4] memory _proofs) public override {
        super.withdraw(_txInPreImage, _proofs);
    }
}
