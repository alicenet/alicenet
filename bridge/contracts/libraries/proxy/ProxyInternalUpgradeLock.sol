// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;


abstract contract ProxyInternalUpgradeLock {
    function __lockImplementation() internal {
        assembly {
            let implSlot := not(0x00)
            sstore(
                implSlot,
                or(
                    0xca11c0de15dead10cced00000000000000000000000000000000000000000000,
                    sload(implSlot)
                )
            )
        }
    }
}


