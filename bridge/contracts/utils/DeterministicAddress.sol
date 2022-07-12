// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

abstract contract DeterministicAddress {
    function getMetamorphicContractAddress(bytes32 _salt, address _factory)
        public
        pure
        returns (address)
    {
        // byte code for metamorphic contract
        // 6020363636335afa1536363636515af43d36363e3d36f3
        bytes32 metamorphicContractBytecodeHash_ = 0x1c0bf703a3415cada9785e89e9d70314c3111ae7d8e04f33bb42eb1d264088be;
        return
            address(
                uint160(
                    uint256(
                        keccak256(
                            abi.encodePacked(
                                hex"ff",
                                _factory,
                                _salt,
                                metamorphicContractBytecodeHash_
                            )
                        )
                    )
                )
            );
    }

    function getStaticPoolContractAddress(bytes32 _salt, address _bridgeRouter)
        public
        pure
        returns (address)
    {
        // does not work: 5880818283335afa3d82833e3d91f3
        // bytes32 metamorphicContractBytecodeHash_ = 0xcd77112ba3315c30f6863dae90cb281bf2f644ef3fd9d21e53d1968182daa472;

        // works: 5880818283335afa3d82833e3d82f3
        bytes32 metamorphicContractBytecodeHash_ = 0xf231e946a2f88d89eafa7b43271c54f58277304b93ac77d138d9b0bb3a989b6d;
        return
            address(
                uint160(
                    uint256(
                        keccak256(
                            abi.encodePacked(
                                hex"ff",
                                _bridgeRouter,
                                _salt,
                                metamorphicContractBytecodeHash_
                            )
                        )
                    )
                )
            );
    }
}
