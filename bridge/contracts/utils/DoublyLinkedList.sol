// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.0;

contract DoublyLinkedList {
    struct Object {
        uint64 id;
        uint64 next;
        uint64 prev;
        address data;
    }

    uint256 public head;
    uint256 public tail;
    uint256 public idCounter;
    mapping(uint256 => Object) public objects;

    /**
     * @dev Creates an empty list.
     */
    constructor() {
        head = 0;
        tail = 0;
        idCounter = 1;
    }

    /**
     * @dev Retrieves the Object denoted by `_id`.
     */
    function get(uint256 _id) public view virtual returns (Object memory) {
        return objects[_id];
    }

    /**
     * @dev Retrieves the Object denoted by `_id`.
     */
    function getValue(uint256 _id) public view virtual returns (bytes memory) {
        return abi.encode(objects[_id].data);
    }

    /**
     * @dev Retrieves the Object denoted by `_id`.
     */
    function getValueStruct(uint256 _id) public view virtual returns (Value memory) {
        return objects[_id].data;
    }

    /**
     * @dev Insert a new Object as the new Head with `_data` in the data field.
     */
    function addHead(bytes memory _data) public virtual {
        Value memory dataDecoded = abi.decode(_data, (Value));
        uint256 objectId = _createObject(dataDecoded);
        _link(objectId, head);
        _setHead(objectId);
        if (tail == 0) _setTail(objectId);
    }

    /**
     * @dev Insert a new Object as the new Tail with `_data` in the data field.
     */
    function addTail(bytes memory _data) public virtual {
        if (head == 0) {
            addHead(_data);
        } else {
            Value memory dataDecoded = abi.decode(_data, (Value));
            uint256 objectId = _createObject(dataDecoded);
            _link(tail, objectId);
            _setTail(objectId);
        }
    }

    /**
     * @dev Remove the Object denoted by `_id` from the List.
     */
    function remove(uint256 _id) public virtual {
        Object memory removeObject = objects[_id];
        if (head == _id && tail == _id) {
            _setHead(0);
            _setTail(0);
        } else if (head == _id) {
            _setHead(removeObject.next);
            objects[removeObject.next].prev = 0;
        } else if (tail == _id) {
            _setTail(removeObject.prev);
            objects[removeObject.prev].next = 0;
        } else {
            _link(removeObject.prev, removeObject.next);
        }
        delete objects[removeObject.id];
    }

    /**
     * @dev Insert a new Object after the Object denoted by `_id` with `_data` in the data field.
     */
    function insertAfter(uint256 _prevId, bytes memory _data) public virtual {
        if (_prevId == tail) {
            addTail(_data);
        } else {
            Object memory prevObject = objects[_prevId];
            Object memory nextObject = objects[prevObject.next];
            Value memory dataDecoded = abi.decode(_data, (Value));
            uint256 newObjectId = _createObject(dataDecoded);
            _link(newObjectId, nextObject.id);
            _link(prevObject.id, newObjectId);
        }
    }

    /**
     * @dev Insert a new Object before the Object denoted by `_id` with `_data` in the data field.
     */
    function insertBefore(uint256 _nextId, bytes memory _data) public virtual {
        if (_nextId == head) {
            addHead(_data);
        } else {
            insertAfter(objects[_nextId].prev, _data);
        }
    }

    /**
     * @dev Internal function to update the Head pointer.
     */
    function _setHead(uint256 _id) internal {
        head = _id;
    }

    /**
     * @dev Internal function to update the Tail pointer.
     */
    function _setTail(uint256 _id) internal {
        tail = _id;
    }

    /**
     * @dev Internal function to create an unlinked Object.
     */
    function _createObject(Value memory _data) internal returns (uint256) {
        uint256 newId = idCounter;
        idCounter += 1;
        Object memory object = Object(newId, 0, 0, _data);
        objects[object.id] = object;
        return object.id;
    }

    /**
     * @dev Internal function to link an Object to another.
     */
    function _link(uint256 _prevId, uint256 _nextId) internal {
        objects[_prevId].next = _nextId;
        objects[_nextId].prev = _prevId;
    }
}
