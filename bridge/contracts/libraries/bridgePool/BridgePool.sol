// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

library BridgePool {
    // @notice calculates salt for a BridgePool contract based on ERC contract's address, tokenType, chainID and version_
    // @param tokenContractAddr_ address of ERC contract of BridgePool
    // @param tokenType_ type of token (1=ERC20, 2=ERC721)
    // @param version_ version of the implementation
    // @param chainID_ chain ID
    // @return calculated calculated salt
    function getBridgePoolSalt(
        address tokenContractAddr_,
        uint8 tokenType_,
        uint256 chainID_,
        uint16 version_
    ) public pure returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked(tokenContractAddr_)),
                    keccak256(abi.encodePacked(tokenType_)),
                    keccak256(abi.encodePacked(chainID_)),
                    keccak256(abi.encodePacked(version_))
                )
            );
    }
}
