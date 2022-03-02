// SPDX-License-Identifier: MIT-open-group
pragma solidity  ^0.8.11;
import "contracts/utils/DeterministicAddress.sol";
import "contracts/libraries/proxy/ProxyUpgrader.sol";
contract MockFactory is DeterministicAddress, ProxyUpgrader {

    /**
    @dev owner role for priveledged access to functions
    */
    address private owner_;

    /**
    @dev delegator role for priveledged access to delegateCallAny
    */
    address private delegator_;

    /**
    @dev array to store list of contract salts
    */
    bytes32[] private contracts_;

    /**
    @dev slot for storing implementation address
    */
    address private implementation_;
    function setOwner(address _new) public {
        owner_ = _new;
    }
}