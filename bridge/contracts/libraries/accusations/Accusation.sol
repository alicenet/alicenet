// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/utils/AccusationsLibrary.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/utils/DeterministicAddress.sol";
import "contracts/libraries/errorCodes/AccusationsErrorCodes.sol";

abstract contract Accusation is DeterministicAddress, ImmutableFactory {
    bytes32 internal constant _ACCUSATION_TAG = "Accusation";
    bytes32 public immutable _name; // immutable string not possible

    modifier onlyAccusation() {
        bytes32 salt = getSaltForAccusation(_name);
        require(
            getMetamorphicContractAddress(salt, _factoryAddress()) == msg.sender,
            string(abi.encodePacked(AccusationsErrorCodes.ACCUSATIONS_ONLY_ACCUSATION))
        );
        _;
    }

    constructor(bytes32 name_) {
        _name = name_;
    }

    /**
     * @notice getSaltForAccusation calculates salt for an Accusation contract based on its name and type identifier
     * @param accusationName_ the name of the accusation
     * @return calculated salt
     */
    function getSaltForAccusation(bytes32 accusationName_) public view returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked(accusationName_)),
                    keccak256(abi.encodePacked(_ACCUSATION_TAG))
                )
            );
    }
}
