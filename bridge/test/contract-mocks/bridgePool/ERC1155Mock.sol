// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC1155/ERC1155.sol";

contract ERC1155Mock is ERC1155 {

    uint256 public constant FUNGIBLE_1 = 0;
    uint256 public constant FUNGIBLE_2 = 1;
    uint256 public constant NON_FUNGIBLE_1 = 2;
    uint256 public constant NON_FUNGIBLE_2 = 3;

    constructor() ERC1155("https://example.com/api/item/{id}.json") {
        _mint(msg.sender, FUNGIBLE_1, 10**18, "");
        _mint(msg.sender, FUNGIBLE_2, 10**27, "");
        _mint(msg.sender, NON_FUNGIBLE_1, 1, "");
        _mint(msg.sender, NON_FUNGIBLE_2, 1, "");
    }

    function mint(address to, uint256 tokenId, uint256 tokenAmount) external {
        _mint(to, tokenId, tokenAmount,"");
    }


}
