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
        uint32 prevEpoch,
        uint32 nextEpoch
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

library DoublyLinkedLists {

    using NodeUpdate for Node;

    /**
     * @dev Retrieves the Node denoted by `epoch`.
     */
    function getNode(DoublyLinkedList storage list, uint256 epoch) public view returns (Node memory) {
        return list.nodes[epoch];
    }

    /**
     * @dev Retrieves the Node value denoted by `epoch`.
     */
    function getValue(DoublyLinkedList storage list, uint256 epoch) public view returns (address) {
        return list.nodes[epoch].data;
    }


    /**
     * @dev Retrieves the next epoch of a Node denoted by `epoch`.
     */
    function getNextEpoch(DoublyLinkedList storage list, uint256 epoch) public view returns (uint256) {
        return list.nodes[epoch].next;
    }

    /**
     * @dev Checks if a node is inserted into the list at the specified `epoch`.
     */
    function exists(DoublyLinkedList storage list, uint256 epoch) public view returns (bool) {
        return _exists(list, epoch);
    }

    /**
     * @dev Function to update the Head pointer.
     */
    function setHead(DoublyLinkedList storage list, uint256 epoch) public {
        _setHead(list, epoch);
    }

    /**
     * @dev Insert a new Node in the `epoch` position with `data` address in the data field.
            This function fails if epoch is smaller or equal than the current head, if
            the data field is the zero address or if a node already exists at `epoch`.
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
        if (_exists(list, epoch)) {
            revert DoublyLinkedListErrors.ExistentNodeAtPosition(epoch);
        }
        if (data == address(0)) {
            revert DoublyLinkedListErrors.InvalidData();
        }
        Node memory node = _createNode(epoch, data);
        // initialization case
        if (head == 0) {
            list.nodes[epoch] = node;
            _setHead(list, epoch);
            // if head is 0, then the tail is also 0 and should be also initialized
            _setTail(list, epoch);
            return;
        }
        // appending after the tail
        if (epoch > tail) {
            list.nodes[epoch] = node.updatePrevious(tail);
            _linkNext(list, tail, epoch);
            _setTail(list, epoch);
            return;
        }
        // appending between the tail and the head
        uint32 currentPosition = tail;
        while (currentPosition != head) {
            Node memory currentNode = list.nodes[currentPosition];
            if (epoch > currentNode.prev) {
                list.nodes[epoch] = node.update(currentNode.prev, currentNode.epoch);
               _linkNext(list, currentNode.prev, epoch);
               _linkPrevious(list, currentNode.epoch, epoch);
               return;
            }
            currentPosition = currentNode.prev;
        }
        // should not be possible to reach this point
        revert DoublyLinkedListErrors.InvalidNodeInsertion(head, tail, epoch);
    }

    // Internal function to check if a node is inserted into the list at the specified `epoch`.
     function _exists(DoublyLinkedList storage list, uint256 epoch) public view returns (bool) {
        return list.nodes[epoch].data != address(0);
    }

    // Internal function to update the Head pointer.
    function _setHead(DoublyLinkedList storage list, uint256 epoch) internal {
        list.head = epoch;
    }

    // Internal function to update the Tail pointer.
    function _setTail(DoublyLinkedList storage list, uint256 epoch) internal {
        list.tail = epoch;
    }

    // Internal function to create a new node Object.
    function _createNode(uint32 epoch, address data) internal pure returns (Node memory) {
        return Node(epoch, 0, 0, data);
    }

    // Internal function to link an Node to its next node.
    function _linkNext(
        DoublyLinkedList storage list,
        uint32 prevEpoch,
        uint32 nextEpoch
    ) internal {
        list.nodes[prevEpoch].next = nextEpoch;
    }

    // Internal function to link an Node to its previous node.
    function _linkPrevious(
        DoublyLinkedList storage list,
        uint32 nextEpoch,
        uint32 prevEpoch
    ) internal {
        list.nodes[nextEpoch].prev = prevEpoch;
    }
}
