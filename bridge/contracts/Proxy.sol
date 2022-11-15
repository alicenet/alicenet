// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/libraries/proxy/ProxyInternalUpgradeLock.sol";
import "contracts/utils/DeterministicAddress.sol";

/**
 *@dev RUN OPTIMIZER OFF
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
contract Proxy is ProxyInternalUpgradeLock {
    address private immutable _factory;

    modifier onlyFactory() {
        require(msg.sender == _factory, "onlyFactory");
        _;
    }

    constructor() {
        _factory = msg.sender;
    }

    receive() external payable {
        _fallback();
    }

    fallback() external payable {
        _fallback();
    }

    /// Returns the implementation address (target) of the Proxy
    /// @return implAddr the implementation address
    function getImplementationAddress() public view returns (address implAddr) {
        assembly ("memory-safe") {
            implAddr := and(
                sload(not(0x00)),
                0x000000000000000000000000ffffffffffffffffffffffffffffffffffffffff
            )
        }
    }

    /// @notice changes the implementation address that a proxy is pointed to
    /// @param implementationAddress_ address of the logic
    function setImplementationAddress(address implementationAddress_) public onlyFactory {
        assembly ("memory-safe") {
            // check if the proxy is locked
            if eq(shr(160, sload(not(0x00))), 0xca11c0de15dead10cced0000) {
                mstore(0x00, "imploc")
                revert(0x00, 0x20)
            }
            // store address into slot
            sstore(not(0x00), implementationAddress_)
        }
    }

    /// Delegates calls to proxy implementation
    function _fallback() internal {
        assembly ("memory-safe") {
            let logicAddress := sload(not(0x00))
            if iszero(logicAddress) {
                mstore(0x00, "logicNotSet")
                revert(0x00, 0x20)
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
    }
}
