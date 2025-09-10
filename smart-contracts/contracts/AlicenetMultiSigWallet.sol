// SPDX-License-Identifier: LGPL-3.0-only
pragma solidity >=0.7.0 <0.9.0;

import "@safe-global/safe-contracts/contracts/SafeL2.sol";

contract AlicenetMultiSigWallet is SafeL2 {
    constructor() {
        threshold = 0;
    }
}
