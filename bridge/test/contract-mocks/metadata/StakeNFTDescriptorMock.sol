// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/libraries/metadata/StakeNFTDescriptor.sol";

contract StakeNFTDescriptorMock {
    function constructTokenURI(StakeNFTDescriptor.ConstructTokenURIParams memory params)
        public
        pure
        returns (string memory)
    {
        return StakeNFTDescriptor.constructTokenURI(params);
    }
}
