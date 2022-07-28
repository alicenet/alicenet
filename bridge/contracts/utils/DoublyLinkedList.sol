// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

import {DoublyLinkedListErrors} from "contracts/libraries/errors/DoublyLinkedListErrors.sol";

struct Node {
    uint32 epoch;
    uint32 next;
    uint32 prev;
    address data;
}

library NodeUpdate {
    function update(
        Node memory node,
        uint32 nextEpoch,
        uint32 prevEpoch
    ) internal pure returns (Node memory) {
        node.prev = prevEpoch;
        node.next = nextEpoch;
        return node;
    }

    function updatePrevious(
        Node memory node,
        uint32 prevEpoch
    ) internal pure returns (Node memory) {
        node.prev = prevEpoch;
        return node;
    }

    function updateNext(
        Node memory node,
        uint32 nextEpoch
    ) internal pure returns (Node memory) {
        node.next = nextEpoch;
        return node;
    }
}

struct DoublyLinkedList {
    uint256 head;
    uint256 tail;
    mapping(uint256 => Node) nodes;
}

contract DynamicValues {
    DoublyLinkedList internal _list;
}

library DoublyLinkedLists {

    using NodeUpdate for Node;
    /**
     * @dev Retrieves the Object denoted by `_id`.
     */
    function get(DoublyLinkedList storage list, uint256 id) public view returns (Node memory) {
        return list.nodes[id];
    }

    /**
     * @dev Retrieves the Node value denoted by `_id`.
     */
    function getValue(DoublyLinkedList storage list, uint256 id) public view returns (address) {
        return list.nodes[id].data;
    }

    /**
     * @dev Insert a new Node in the `id` position with `data` address in the data field.
            This function fails if id is smaller or equal than the current head.
     */
    function addNode(
        DoublyLinkedList storage list,
        uint32 epoch,
        address data
    ) public {
        uint32 head = uint32(list.head);
        uint32 tail = uint32(list.tail);
        if (epoch <= head) {
            revert DoublyLinkedListErrors.InvalidNodeId(head, tail, epoch);
        }
        Node memory node = _createNode(epoch, data);
        if (head == 0) {
            list.nodes[epoch] = node;
            _setHead(list, node.epoch);
            if (tail == 0) {
                _setTail(list, node.epoch);
            }
            return;
        }
        if (node.epoch >= tail) {
            list.nodes[epoch] = node.updatePrevious(tail);
            _linkNext(list, tail, node.epoch);
            _setTail(list, node.epoch);
            return;
        }
        uint32 currentPosition = tail;
        while (currentPosition != head) {
            Node memory currentNode = list.nodes[currentPosition];
            if (node.epoch >= currentNode.prev) {
                list.nodes[epoch] = node.update(currentNode.epoch, currentNode.prev);
               _linkNext(list, currentNode.prev, node.epoch);
               _linkPrevious(list, currentNode.epoch, node.epoch);
            }
            currentPosition = currentNode.prev;
        }

    }

    /**
     * @dev Internal function to update the Head pointer.
     */
    function _setHead(DoublyLinkedList storage list, uint256 id) internal {
        list.head = id;
    }

    /**
     * @dev Internal function to update the Tail pointer.
     */
    function _setTail(DoublyLinkedList storage list, uint256 id) internal {
        list.tail = id;
    }

    function _createNode(uint32 epoch, address data) internal pure returns (Node memory) {
        return Node(epoch, 0, 0, data);
    }

    /**
     * @dev Internal function to link an Node to another.
     */
    function _linkNext(
        DoublyLinkedList storage list,
        uint32 prevEpoch,
        uint32 nextEpoch
    ) internal {
        list.nodes[prevEpoch].next = nextEpoch;
    }

    function _linkPrevious(
        DoublyLinkedList storage list,
        uint32 nextEpoch,
        uint32 prevEpoch
    ) internal {
        list.nodes[nextEpoch].prev = prevEpoch;
    }
}
