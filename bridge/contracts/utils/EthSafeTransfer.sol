// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;


abstract contract EthSafeTransfer {

    /// @notice _safeTransferEth performs a transfer of Eth using the call
    /// method / this function is resistant to breaking gas price changes and /
    /// performs call in a safe manner by reverting on failure. / this function
    /// will return without performing a call or reverting, / if amount_ is zero
    function _safeTransferEth(address to_, uint256 amount_) internal {
        if (amount_ == 0 ){
            return;
        }
        require(to_ != address(0), "EthSafeTransfer: cannot transfer ETH to address 0x0");
        address payable caller = payable(to_);
        (bool success, ) = caller.call{value: amount_}("");
        require(success, "EthSafeTransfer: Transfer failed.");
    }

}