// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/DeterministicAddress.sol";

/**
 *@notice RUN OPTIMIZER OFF
 */
/**
 * @notice Proxy is a delegatecall reverse proxy implementation
 * the forwarding address is stored at the slot location of not(0)
 * if not(0) has a value stored in it that is of the form 0Xca11c0de15dead10cced0000< address >
 * the proxy may no longer be upgraded using the internal mechanism. This does not prevent the implementation
 * from upgrading the proxy by changing this slot.
 * The proxy may be directly upgraded ( if the lock is not set )
 * by calling the proxy from the factory address using the format
 * abi.encodeWithSelector(0xca11c0de, <address>);
 * All other calls will be proxied through to the implementation.
 * The implementation can not be locked using the internal upgrade mechanism due to the fact that the internal
 * mechanism zeros out the higher order bits. Therefore, the implementation itself must carry the locking mechanism that sets
 * the higher order bits to lock the upgrade capability of the proxy.
 */
contract Proxy {
    address private immutable _factory;

    constructor() {
        _factory = msg.sender;
    }

    receive() external payable {
        _fallback();
    }

    fallback() external payable {
        _fallback();
    }

    /// Delegates calls to proxy implementation
    function _fallback() internal {
        bytes memory errorString = abi.encodeWithSelector(0x08c379a0, "imploc");
        // make local copy of factory since immutables
        // are not accessable in assembly as of yet
        address factory = _factory;
        assembly ("memory-safe") {
            function getImplementationAddress() -> implAddr {
                implAddr := and(
                    sload(not(0x00)),
                    0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff
                )
            }

            // admin is the builtin logic to change the implementation
            function admin() {
                // this is an assignment to implementation
                let newImpl := shr(96, shl(96, calldataload(0x04)))
                if eq(shr(160, sload(not(0x00))), 0x00ca11c0de15dead10cced00) {
                    let ptr:= mload(0x40)
                    let basePtr := ptr
                    mstore(ptr, hex"08c379a0")
                    mstore(add(ptr, 0x4), 0x20)
                    //num characters 0x6 = 6
                    mstore(add(ptr, 0x24), 0x6)
                    ptr := add(ptr, 0x44)
                    mstore(ptr, "imploc")
                    ptr := add(ptr, 0x20)
                    revert(basePtr, sub(ptr, basePtr))
                    
                }
                // store address into slot
                sstore(not(0x00), newImpl)
                stop()
            }

            // passthrough is the passthrough logic to delegate to the implementation
            function passthrough() {
                let logicAddress := sload(not(0x00))
                if iszero(logicAddress) {
                    let ptr:= mload(0x40)
                    let basePtr := ptr
                    mstore(ptr, hex"08c379a0")
                    mstore(add(ptr, 0x4), 0x20)
                    //num characters 0xb = 11
                    mstore(add(ptr, 0x24), 0xb)
                    ptr := add(ptr, 0x44)
                    mstore(ptr, "logicNotSet")
                    ptr := add(ptr, 0x20)
                    revert(basePtr, sub(ptr, basePtr))
                }
                // load free memory pointer
                let _ptr := mload(0x40)
                // allocate memory proportionate to calldata
                mstore(0x40, add(_ptr, calldatasize()))
                // copy calldata into memory
                calldatacopy(_ptr, 0x00, calldatasize())
                let ret := delegatecall(gas(), logicAddress, _ptr, calldatasize(), 0x00, 0x00)
                returndatacopy(_ptr, 0x00, returndatasize())
                if iszero(ret) {
                    revert(_ptr, returndatasize())
                }
                return(_ptr, returndatasize())
            }
            if or(eq(calldatasize(), 0x24), eq(calldatasize(), 0x5)) {
                {
                    let selector := shr(216, calldataload(0x00))
                    switch selector
                    case 0x0cbcae703c {
                        // getImplementationAddress()
                        let ptr := mload(0x40)
                        let retAddress := getImplementationAddress()
                        mstore(ptr, retAddress)
                        return(ptr, 0x20)
                    }
                    case 0xca11c0de00 {
                        // if caller is factory,
                        // and has 0xca11c0de<address> as calldata
                        // run admin logic and return
                        if eq(caller(), factory) {
                            admin()
                        }
                    }
                    default {
                        mstore(0x00, "function not found")
                        revert(0x00, 0x20)
                    }
                }
            }
            // admin logic was not run so fallthrough to delegatecall
            passthrough()
        }
    }
}

// 0000000000000000000000000000000000000000000000000000000000000000
// 0000000000000000000000000000000000000000000000000000000000000000
// 0000000000000000000000000000000000000000000000000000000000000080
// 0000000000000000000000000000000000000000000000000000000000000000
// 08c379a0000000000000000000000000000000000000000000000000000000000000002000000000000000000000000000000000000000000000000000000000000000066861686168610000000000000000000000000000000000000000000000000000
// 0x5da7d2e0
696d706c6f630