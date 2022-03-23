// SPDX-License-Identifier: GPL-2.0-or-later
pragma solidity >=0.5.0;

import "contracts/interfaces/INFTStake.sol";

/// @title Describes a staked position NFT tokens via URI
interface INFTStakeDescriptor {
    /// @notice Produces the URI describing a particular token ID for a staked position
    /// @dev Note this URI may be a data: URI with the JSON contents directly inlined
    /// @param _stakeNft The stake NFT for which to describe the token
    /// @param tokenId The ID of the token for which to produce a description, which may not be valid
    /// @return The URI of the ERC721-compliant metadata
    function tokenURI(INFTStake _stakeNft, uint256 tokenId) external view returns (string memory);
}
