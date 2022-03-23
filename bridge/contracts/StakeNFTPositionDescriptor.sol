// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity ^0.8.11;
pragma abicoder v2;

import "contracts/libraries/metadata/StakeNFTDescriptor.sol";
import "contracts/interfaces/INFTStakeDescriptor.sol";

/// @custom:salt StakeNFTPositionDescriptor
/// @custom:deploy-type deployUpgradeable
contract StakeNFTPositionDescriptor is INFTStakeDescriptor {
    function tokenURI(INFTStake _stakeNft, uint256 tokenId)
        external
        view
        override
        returns (string memory)
    {
        (
            uint256 shares,
            uint256 freeAfter,
            uint256 withdrawFreeAfter,
            uint256 accumulatorEth,
            uint256 accumulatorToken
        ) = _stakeNft.getPosition(tokenId);

        return
            StakeNFTDescriptor.constructTokenURI(
                StakeNFTDescriptor.ConstructTokenURIParams({
                    tokenId: tokenId,
                    shares: shares,
                    freeAfter: freeAfter,
                    withdrawFreeAfter: withdrawFreeAfter,
                    accumulatorEth: accumulatorEth,
                    accumulatorToken: accumulatorToken
                })
            );
    }
}
