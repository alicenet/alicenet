// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library BaseParserLibrary {
    // Size of a word, in bytes.
    uint256 internal constant WORD_SIZE = 32;
    // Size of the header of a 'bytes' array.
    uint256 internal constant BYTES_HEADER_SIZE = 32;

    /// @notice Extracts a uint32 from a little endian bytes array.
    /// @param src the binary data
    /// @param offset place inside `src` to start reading data from
    /// @return val a uint32
    /// @dev ~559 gas
    function extractUInt32(bytes memory src, uint256 offset)
        internal
        pure
        returns (uint32 val)
    {
        require(
            offset + 4 > offset,
            "BaseParserLibrary: An overflow happened with the offset parameter!"
        );
        require(
            src.length >= offset + 4,
            "BaseParserLibrary: Trying to read an offset out of boundaries in the src binary!"
        );

        assembly {
            val := shr(sub(256, 32), mload(add(add(src, 0x20), offset)))
            val := or(
                or(
                    or(
                        shr(24, and(val, 0xff000000)),
                        shr(8, and(val, 0x00ff0000))
                    ),
                    shl(8, and(val, 0x0000ff00))
                ),
                shl(24, and(val, 0x000000ff))
            )
        }
    }

    /// @notice Extracts a uint16 from a little endian bytes array.
    /// @param src the binary data
    /// @param offset place inside `src` to start reading data from
    /// @return val a uint16
    /// @dev ~204 gas
    function extractUInt16(bytes memory src, uint256 offset)
        internal
        pure
        returns (uint16 val)
    {
        require(
            offset + 2 > offset,
            "BaseParserLibrary: Error extracting uin16! An overflow happened with the offset parameter!"
        );
        require(
            src.length >= offset + 2,
            "BaseParserLibrary: Error extracting uin16! Trying to read an offset out of boundaries in the src binary!"
        );

        assembly {
            val := shr(sub(256, 16), mload(add(add(src, 0x20), offset)))
            val := or(
                        shr(8, and(val, 0xff00)),
                        shl(8, and(val, 0x00ff))
            )
        }
    }

    /// @notice Extracts a uint16 from a big endian bytes array.
    /// @param src the binary data
    /// @param offset place inside `src` to start reading data from
    /// @return val a uint16
    /// @dev ~204 gas
    function extractUInt16FromBigEndian(bytes memory src, uint256 offset)
        internal
        pure
        returns (uint16 val)
    {
        require(
            offset + 2 > offset,
            "BaseParserLibrary: Error extracting uin16! An overflow happened with the offset parameter!"
        );
        require(
            src.length >= offset + 2,
            "BaseParserLibrary: Error extracting uin16! Trying to read an offset out of boundaries in the src binary!"
        );

        assembly {
            val := and(shr(sub(256, 16), mload(add(add(src, 0x20), offset))), 0xffff)
        }
    }

    /// @notice Extracts a bool from a bytes array.
    /// @param src the binary data
    /// @param offset place inside `src` to start reading data from
    /// @return a bool
    /// @dev ~204 gas
    function extractBool(bytes memory src, uint256 offset)
        internal
        pure
        returns (bool)
    {
        require(
            offset + 1 > offset,
            "BaseParserLibrary: Error extracting bool! An overflow happened with the offset parameter!"
        );
        require(
            src.length >= offset + 1,
            "BaseParserLibrary: Error extracting bool! Trying to read an offset out of boundaries in the src binary!"
        );
        uint256 val;
        assembly {
            val := shr(sub(256, 8), mload(add(add(src, 0x20), offset)))
            val := and(val, 0x01)
        }
        return val == 1;
    }

    /// @notice Extracts a uint256 from a little endian bytes array.
    /// @param src the binary data
    /// @param offset place inside `src` to start reading data from
    /// @return val a uint256
    /// @dev ~5155 gas
    function extractUInt256(bytes memory src, uint256 offset)
        internal
        pure
        returns (uint256 val)
    {
        require(
            offset + 31 > offset,
            "BaseParserLibrary: An overflow happened with the offset parameter!"
        );
        require(
            src.length > offset + 31,
            "BaseParserLibrary: Trying to read an offset out of boundaries!"
        );

        assembly {
            val := mload(add(add(src, 0x20), offset))
        }
    }

    /// @notice Extracts a uint256 from a big endian bytes array.
    /// @param src the binary data
    /// @param offset place inside `src` to start reading data from
    /// @return val a uint256
    /// @dev ~1400 gas
    function extractUInt256FromBigEndian(bytes memory src, uint256 offset)
        internal
        pure
        returns (uint256 val)
    {
        require(
            offset + 31 > offset,
            "BaseParserLibrary: An overflow happened with the offset parameter!"
        );
        require(
            src.length > offset + 31,
            "BaseParserLibrary: Trying to read an offset out of boundaries!"
        );

        uint256 srcDataPointer;
        uint32 val0 = 0;
        uint32 val1 = 0;
        uint32 val2 = 0;
        uint32 val3 = 0;
        uint32 val4 = 0;
        uint32 val5 = 0;
        uint32 val6 = 0;
        uint32 val7 = 0;

        assembly {
            srcDataPointer := mload(add(add(src, 0x20), offset))
            val0 := and(srcDataPointer, 0xffffffff)
            val1 := and(shr(32, srcDataPointer), 0xffffffff)
            val2 := and(shr(64, srcDataPointer), 0xffffffff)
            val3 := and(shr(96, srcDataPointer), 0xffffffff)
            val4 := and(shr(128, srcDataPointer), 0xffffffff)
            val5 := and(shr(160, srcDataPointer), 0xffffffff)
            val6 := and(shr(192, srcDataPointer), 0xffffffff)
            val7 := and(shr(224, srcDataPointer), 0xffffffff)

            val0 := or(
                or(
                    or(
                        shr(24, and(val0, 0xff000000)),
                        shr(8, and(val0, 0x00ff0000))
                    ),
                    shl(8, and(val0, 0x0000ff00))
                ),
                shl(24, and(val0, 0x000000ff))
            )
            val1 := or(
                or(
                    or(
                        shr(24, and(val1, 0xff000000)),
                        shr(8, and(val1, 0x00ff0000))
                    ),
                    shl(8, and(val1, 0x0000ff00))
                ),
                shl(24, and(val1, 0x000000ff))
            )
            val2 := or(
                or(
                    or(
                        shr(24, and(val2, 0xff000000)),
                        shr(8, and(val2, 0x00ff0000))
                    ),
                    shl(8, and(val2, 0x0000ff00))
                ),
                shl(24, and(val2, 0x000000ff))
            )
            val3 := or(
                or(
                    or(
                        shr(24, and(val3, 0xff000000)),
                        shr(8, and(val3, 0x00ff0000))
                    ),
                    shl(8, and(val3, 0x0000ff00))
                ),
                shl(24, and(val3, 0x000000ff))
            )
            val4 := or(
                or(
                    or(
                        shr(24, and(val4, 0xff000000)),
                        shr(8, and(val4, 0x00ff0000))
                    ),
                    shl(8, and(val4, 0x0000ff00))
                ),
                shl(24, and(val4, 0x000000ff))
            )
            val5 := or(
                or(
                    or(
                        shr(24, and(val5, 0xff000000)),
                        shr(8, and(val5, 0x00ff0000))
                    ),
                    shl(8, and(val5, 0x0000ff00))
                ),
                shl(24, and(val5, 0x000000ff))
            )
            val6 := or(
                or(
                    or(
                        shr(24, and(val6, 0xff000000)),
                        shr(8, and(val6, 0x00ff0000))
                    ),
                    shl(8, and(val6, 0x0000ff00))
                ),
                shl(24, and(val6, 0x000000ff))
            )
            val7 := or(
                or(
                    or(
                        shr(24, and(val7, 0xff000000)),
                        shr(8, and(val7, 0x00ff0000))
                    ),
                    shl(8, and(val7, 0x0000ff00))
                ),
                shl(24, and(val7, 0x000000ff))
            )

            val :=
            or(
                or(
                    or(
                        or(
                            or(
                                or(
                                    or(
                                        shl(224, val0),
                                        shl(192, val1)
                                    ),
                                    shl(160, val2)
                                ),
                                shl(128, val3)
                            ),
                            shl(96, val4)
                        ),
                        shl(64, val5)
                    ),
                    shl(32, val6)
                ),
                val7
            )
        }
    }

    /// @notice Reverts a bytes array. Can be used to convert an array from little endian to big endian and vice-versa.
    /// @param orig the binary data
    /// @return reversed the reverted bytes array
    /// @dev ~13832 gas
    function reverse(bytes memory orig)
        internal
        pure
        returns (bytes memory reversed)
    {
        reversed = new bytes(orig.length);
        for (uint256 idx = 0; idx < orig.length; idx++) {
            reversed[orig.length - idx - 1] = orig[idx];
        }
    }

    /// @notice Copy 'len' bytes from memory address 'src', to address 'dest'. This function does not check the or destination, it only copies the bytes.
    /// @param src the pointer to the source
    /// @param dest the pointer to the destination
    /// @param len the len of data to be copied
    function copy(
        uint256 src,
        uint256 dest,
        uint256 len
    ) internal pure {
        // Copy word-length chunks while possible
        for (; len >= WORD_SIZE; len -= WORD_SIZE) {
            assembly {
                mstore(dest, mload(src))
            }
            dest += WORD_SIZE;
            src += WORD_SIZE;
        }
        // Returning earlier if there's no leftover bytes to copy
        if (len == 0) {
            return;
        }
        // Copy remaining bytes
        uint256 mask = 256**(WORD_SIZE - len) - 1;
        assembly {
            let srcpart := and(mload(src), not(mask))
            let destpart := and(mload(dest), mask)
            mstore(dest, or(destpart, srcpart))
        }
    }

    /// @notice Returns a memory pointer to the data portion of the provided bytes array.
    /// @param bts the bytes array to get a pointer from
    /// @return addr the pointer to the `bts` bytes array
    function dataPtr(bytes memory bts) internal pure returns (uint256 addr) {
        assembly {
            addr := add(bts, BYTES_HEADER_SIZE)
        }
    }

    /// @notice Extracts a bytes array with length `howManyBytes` from `src`'s `offset` forward.
    /// @param src the bytes array to extract from
    /// @param offset where to start extracting from
    /// @param howManyBytes how many bytes we want to extract from `src`
    /// @return out the extracted bytes array
    /// @dev Extracting the 32-64th bytes out of a 64 bytes array takes ~7828 gas.
    function extractBytes(
        bytes memory src,
        uint256 offset,
        uint256 howManyBytes
    ) internal pure returns (bytes memory out) {
        require(
            offset + howManyBytes >= offset,
            "BaseParserLibrary: An overflow happened with the offset or the howManyBytes parameter!"
        );
        require(
            src.length >= offset + howManyBytes,
            "BaseParserLibrary: Not enough bytes to extract in the src binary"
        );
        out = new bytes(howManyBytes);
        uint256 start;

        assembly {
            start := add(add(src, offset), BYTES_HEADER_SIZE)
        }

        copy(start, dataPtr(out), howManyBytes);
    }

    /// @notice Extracts a bytes32 extracted from `src`'s `offset` forward.
    /// @param src the source bytes array to extract from
    /// @param offset where to start extracting from
    /// @return out the bytes32 data extracted from `src`
    /// @dev ~439 gas
    function extractBytes32(bytes memory src, uint256 offset)
        internal
        pure
        returns (bytes32 out)
    {
        require(
            offset + 32 > offset,
            "BaseParserLibrary: An overflow happened with the offset parameter!"
        );
        require(
            src.length >= (offset + 32),
            "BaseParserLibrary: not enough bytes to extract"
        );
        assembly {
            out := mload(add(add(src, BYTES_HEADER_SIZE), offset))
        }
    }
}
