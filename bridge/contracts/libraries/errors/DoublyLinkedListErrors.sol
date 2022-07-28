// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library DoublyLinkedListErrors {
    error InvalidNodeId(uint256 head, uint256 tail, uint256 id);
}
