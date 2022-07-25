// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

struct Node {
    uint64 id;
    uint64 next;
    uint64 prev;
    address data;
}

struct DoublyLinkedList {
    uint256 head;
    uint256 tail;
    uint64 id;
    mapping(uint256 => Node) nodes;
}

contract DynamicValues {
    DoublyLinkedList internal _list;
}

library DoublyLinkedLists {
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
     * @dev Insert a new Node as the new Head with `data` address in the data field.
     */
    function addHead(DoublyLinkedList storage list, address data) public {
        uint256 nodeId = _createNode(list, data);
        _link(list, nodeId, list.head);
        _setHead(list, nodeId);
        if (list.tail == 0) {
            _setTail(list, nodeId);
        }
    }

    /**
     * @dev Insert a new Node as the new Tail with `data` address in the data field.
     */
    function addTail(DoublyLinkedList storage list, address data) public {
        if (list.head == 0) {
            addHead(list, data);
        } else {
            uint256 nodeId = _createNode(list, data);
            _link(list, list.tail, nodeId);
            _setTail(list, nodeId);
        }
    }

    /**
     * @dev Remove the Node denoted by `_id` from the List.
     */
    function remove(DoublyLinkedList storage list, uint256 id) public {
        // todo: raise error in case it doesn't exist
        Node memory removeObject = get(list, id);
        if (list.head == id && list.tail == id) {
            _setHead(list, 0);
            _setTail(list, 0);
        } else if (list.head == id) {
            _setHead(list, removeObject.next);
            list.nodes[removeObject.next].prev = 0;
        } else if (list.tail == id) {
            _setTail(list, removeObject.prev);
            list.nodes[removeObject.prev].next = 0;
        } else {
            _link(list, removeObject.prev, removeObject.next);
        }
        delete list.nodes[removeObject.id];
    }

    /**
     * @dev Insert a new Node after the Node denoted by `_id` with `_data` in the data field.
     */
    function insertAfter(
        DoublyLinkedList storage list,
        uint256 prevId,
        address data
    ) public {
        if (prevId == list.tail) {
            addTail(list, data);
        } else {
            Node memory prevObject = get(list, prevId);
            Node memory nextObject = get(list, prevObject.next);
            uint256 newObjectId = _createNode(list, data);
            _link(list, newObjectId, nextObject.id);
            _link(list, prevObject.id, newObjectId);
        }
    }

    /**
     * @dev Insert a new Node before the Node denoted by `_id` with `_data` in the data field.
     */
    function insertBefore(
        DoublyLinkedList storage list,
        uint256 nextId,
        address data
    ) public {
        if (nextId == list.head) {
            addHead(list, data);
        } else {
            insertAfter(list, list.nodes[nextId].prev, data);
        }
    }

    /**
     * @dev Internal function to update the Head pointer.
     */
    function _setHead(DoublyLinkedList storage list, uint256 _id) internal {
        list.head = _id;
    }

    /**
     * @dev Internal function to update the Tail pointer.
     */
    function _setTail(DoublyLinkedList storage list, uint256 _id) internal {
        list.tail = _id;
    }

    /**
     * @dev Internal function to create an unlinked Node.
     */
    function _createNode(DoublyLinkedList storage list, address data) internal returns (uint256) {
        list.id++;
        uint256 newId = list.id;
        Node memory object = Node(uint64(newId), 0, 0, data);
        list.nodes[object.id] = object;
        return object.id;
    }

    /**
     * @dev Internal function to link an Node to another.
     */
    function _link(
        DoublyLinkedList storage list,
        uint256 prevId,
        uint256 nextId
    ) internal {
        list.nodes[prevId].next = uint64(nextId);
        list.nodes[nextId].prev = uint64(prevId);
    }
}
