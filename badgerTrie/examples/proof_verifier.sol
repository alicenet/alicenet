pragma solidity ^0.6.4;


contract MerkleProof {
    // Check if the bit at the given 'index' in 'self' is set.
    // Returns:
    //  'true' - if the value of the bit is '1'
    //  'false' - if the value of the bit is '0'
    function bitSet(uint256 self, uint16 index) public pure returns (bool) {
        return (self >> (255 - index)) & 1 == 1;
    }

    function checkProof(
        bytes memory proof,
        bytes32 root,
        bytes32 hash,
        uint256 key,
        uint256 bitset,
        uint256 height,
        bool included,
        uint256 proofKey
    ) public pure returns (bool) {
        bytes32 el;
        bytes32 h = hash;

        bytes32 defaultLeaf = bytes32(
            hex"bc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"
        );

        uint16 proofIdx = 0;

        for (uint256 i = 0; i < height; i++) {
            if (bitSet(bitset, 255 - uint16(i))) {
                proofIdx += 32;
                assembly {
                    el := mload(add(proof, proofIdx))
                }
            } else {
                el = defaultLeaf;
            }

            if (bitSet(key, 255 - uint16(i))) {
                h = keccak256(abi.encodePacked(el, h));
            } else {
                h = keccak256(abi.encodePacked(h, el));
            }
        }

        if (included == false) {
            for (uint256 i = 0; i < 256; i++) {
                if (
                    bitSet(key, 255 - uint16(i)) !=
                    bitSet(proofKey, 255 - uint16(i))
                ) {
                    return false;
                }
            }
        }

        return h == root;
    }
}
