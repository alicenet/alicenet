// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/libraries/proxy/ProxyInternalUpgradeLock.sol";
import "contracts/libraries/proxy/ProxyInternalUpgradeUnlock.sol";

contract MockInitializable is ProxyInternalUpgradeLock, ProxyInternalUpgradeUnlock {
    address factory_;
    uint256 public v;
    uint256 public i;
    address public immutable factoryAddress = 0x0BBf39118fF9dAfDC8407c507068D47572623069;
   /**
     * @dev Indicates that the contract has been initialized.
     */
    bool private _initialized;

    /**
     * @dev Indicates that the contract is in the process of being initialized.
     */
    bool private _initializing;

    /**
     * @dev Modifier to protect an initializer function from being invoked twice.
     */
    modifier initializer() {
        // If the contract is initializing we ignore whether _initialized is set in order to support multiple
        // inheritance patterns, but we only do this in the context of a constructor, because in other contexts the
        // contract may have been reentered.
        require(_initializing ? _isConstructor() : !_initialized, "Initializable: contract is already initialized");

        bool isTopLevelCall = !_initializing;
        if (isTopLevelCall) {
            _initializing = true;
            _initialized = true;
        }

        _;

        if (isTopLevelCall) {
            _initializing = false;
        }
    }

    /**
     * @dev Modifier to protect an initialization function so that it can only be invoked by functions with the
     * {initializer} modifier, directly or indirectly.
     */
    modifier onlyInitializing() {
        require(_initializing, "Initializable: contract is not initializing");
        _;
    }
    function isContract(address account) internal view returns (bool) {
        // This method relies on extcodesize/address.code.length, which returns 0
        // for contracts in construction, since the code is only stored at the end
        // of the constructor execution.

        return account.code.length > 0;
    }

    function _isConstructor() private view returns (bool) {
        return !isContract(address(this));
    }

    function initialize(uint256 _i) public virtual initializer{
        __Mock_init(_i);
    }

    function __Mock_init(uint256 _i) internal onlyInitializing {
        __Mock_init_unchained(_i);
    }

    function __Mock_init_unchained(uint256 _i) internal onlyInitializing {
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