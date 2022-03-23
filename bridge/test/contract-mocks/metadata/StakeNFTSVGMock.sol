// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/libraries/metadata/StakeNFTSVG.sol";

contract StakeNFTSVGMock {
    function generateSVG(StakeNFTSVG.StakeNFTSVGParams memory svgParams)
        public
        pure
        returns (string memory svg)
    {
        return StakeNFTSVG.generateSVG(svgParams);
    }
}
