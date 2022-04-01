// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library BaseParserLibraryErrorCodes {
    // BaseParserLibrary error codes
    uint16 public constant BASEPARSERLIB_OFFSET_PARAMETER_OVERFLOW = 1000; // "BaseParserLibrary: An overflow happened with the offset parameter!"
    uint16 public constant BASEPARSERLIB_OFFSET_OUT_OF_BOUNDS = 1001; // "BaseParserLibrary: Trying to read an offset out of boundaries in the src binary!"
    uint16 public constant BASEPARSERLIB_LE_UINT16_OFFSET_PARAMETER_OVERFLOW = 1002; //  "BaseParserLibrary: Error extracting uin16! An overflow happened with the offset parameter!"
    uint16 public constant BASEPARSERLIB_LE_UINT16_OFFSET_OUT_OF_BOUNDS = 1003; //  "BaseParserLibrary: UINT16 ERROR! Trying to read an offset out of boundaries!"
    uint16 public constant BASEPARSERLIB_BE_UINT16_OFFSET_PARAMETER_OVERFLOW = 1004; // "BaseParserLibrary: UINT16 ERROR! An overflow happened with the offset parameter!"
    uint16 public constant BASEPARSERLIB_BE_UINT16_OFFSET_OUT_OF_BOUNDS = 1005; // "BaseParserLibrary: UINT16 ERROR! Trying to read an offset out of boundaries!"
    uint16 public constant BASEPARSERLIB_BOOL_OFFSET_PARAMETER_OVERFLOW = 1006; // "BaseParserLibrary: BOOL ERROR: OVERFLOW!"
    uint16 public constant BASEPARSERLIB_BOOL_OFFSET_OUT_OF_BOUNDS = 1007; //  "BaseParserLibrary: BOOL ERROR: OFFSET OUT OF BOUNDARIES!"
    uint16 public constant BASEPARSERLIB_LE_UINT256_OFFSET_PARAMETER_OVERFLOW = 1008; //  "BaseParserLibrary: Error extracting uin16! An overflow happened with the offset parameter!"
    uint16 public constant BASEPARSERLIB_LE_UINT256_OFFSET_OUT_OF_BOUNDS = 1009; //  "BaseParserLibrary: UINT16 ERROR! Trying to read an offset out of boundaries!"
    uint16 public constant BASEPARSERLIB_BE_UINT256_OFFSET_PARAMETER_OVERFLOW = 1010; // "BaseParserLibrary: UINT16 ERROR! An overflow happened with the offset parameter!"
    uint16 public constant BASEPARSERLIB_BE_UINT256_OFFSET_OUT_OF_BOUNDS = 1011; // "BaseParserLibrary: UINT16 ERROR! Trying to read an offset out of boundaries!"
    uint16 public constant BASEPARSERLIB_BYTES_OFFSET_PARAMETER_OVERFLOW = 1012; // "BaseParserLibrary: An overflow happened with the offset or the howManyBytes parameter!"
    uint16 public constant BASEPARSERLIB_BYTES_OFFSET_OUT_OF_BOUNDS = 1013; //   "BaseParserLibrary: Not enough bytes to extract in the src binary"
    uint16 public constant BASEPARSERLIB_BYTES32_OFFSET_PARAMETER_OVERFLOW = 1014; // "BaseParserLibrary: An overflow happened with the offset parameter!"
    uint16 public constant BASEPARSERLIB_BYTES32_OFFSET_OUT_OF_BOUNDS = 1015; //   "BaseParserLibrary: not enough bytes to extract"
}
