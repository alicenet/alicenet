// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/legacy/utils/MagicValue.sol";
import "contracts/legacy/interfaces/IMagicEthTransfer.sol";

abstract contract MagicEthTransfer is MagicValue {
    function _safeTransferEthWithMagic(IMagicEthTransfer to_, uint256 amount_) internal {
        to_.depositEth{value: amount_}(_getMagic());
    }
}
