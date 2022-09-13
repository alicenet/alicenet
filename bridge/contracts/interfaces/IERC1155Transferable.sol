// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

interface IERC1155Transferable {
    function safeTransferFrom(
        address from,
        address to,
        uint256 tokenId,
        uint256 amount,
        bytes calldata data
    ) external;
}
