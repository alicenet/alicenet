// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/libraries/proxy/ProxyInternalUpgradeLock.sol";
import "contracts/libraries/proxy/ProxyInternalUpgradeUnlock.sol";

/// @custom:salt Mock
contract MockBaseContract is ProxyInternalUpgradeLock, ProxyInternalUpgradeUnlock {
    address factory_;
    uint256 public v;
    uint256 public immutable i;
    string p;
    constructor(uint256 _i, string memory _p) {
        p= _p;
        i = _i;
        factory_ = msg.sender;
    }

    function setv(uint256 _v) public {
        v = _v;
    }

    function lock() public {
        __lockImplementation();
    }

    function unlock() public {
        __unlockImplementation();
    }

    function setFactory(address _factory) public {
        factory_ = _factory;
    }
    function getFactory() external view returns (address) {
        return factory_;
    }
}
interface IMockBaseContract {
    function v() external returns (uint256);

    function i() external returns (uint256);

    function setv(uint256 _v) external;

    function lock() external;

    function unlock() external;

    function fail() external;
}