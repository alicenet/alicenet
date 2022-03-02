// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;


abstract contract Admin {

    // _admin is a privileged role
    address internal _admin;

    constructor(address admin_) {
        _admin = admin_;
    }

    /// @dev onlyAdmin enforces msg.sender is _admin
    modifier onlyAdmin() {
        require(msg.sender == _admin, "Must be admin");
        _;
    }

    // assigns a new admin may only be called by _admin
    function _setAdmin(address admin_) internal {
        _admin = admin_;
    }

    /// @dev getAdmin returns the current _admin
    function getAdmin() public view returns(address) {
        return _admin;
    }

    /// @dev assigns a new admin may only be called by _admin
    function setAdmin(address admin_) public virtual onlyAdmin {
        _setAdmin(admin_);
    }
}