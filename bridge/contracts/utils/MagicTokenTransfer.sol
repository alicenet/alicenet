// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import "contracts/utils/MagicValue.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/interfaces/IMagicTokenTransfer.sol";
import {
    MagicTokenTransferErrorCodes
} from "contracts/libraries/errorCodes/MagicTokenTransferErrorCodes.sol";

abstract contract MagicTokenTransfer is MagicValue {
    function _safeTransferTokenWithMagic(
        IERC20Transferable token_,
        IMagicTokenTransfer to_,
        uint256 amount_
    ) internal {
        bool success = token_.approve(address(to_), amount_);
        require(success, string(abi.encodePacked(MagicTokenTransferErrorCodes.MAGICTOKENTRANSFER_TRANSFER_FAILED)));
        to_.depositToken(_getMagic(), amount_);
        token_.approve(address(to_), 0);
    }
}
