// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/errors/BridgePoolFactoryErrors.sol";
import "contracts/interfaces/IBridgePool.sol";

abstract contract BridgePoolFactoryBase is ImmutableFactory {
    uint256 private immutable _chainID;
    bool private _publicPoolDeploymentEnabled;
    address private _implementation;

    event BridgePoolCreated(address poolAddress, address token);

    modifier onlyFactoryOrPublicPoolDeploymentEnabled() {
        if (msg.sender != _factoryAddress() && !_publicPoolDeploymentEnabled) {
            revert BridgePoolFactoryErrors.PublicPoolDeploymentTemporallyDisabled();
        }
        _;
    }

    constructor(uint256 chainID_) ImmutableFactory(msg.sender) {
        _chainID = chainID_;
    }

    /**
     * @notice returns bytecode for a Minimal Proxy (EIP-1167) that routes to BridgePool implementation
     */
    fallback() external {
        address implementation_ = _implementation;
        assembly {
            let ptr := mload(0x40)
            mstore(ptr, shl(176, 0x363d3d373d3d3d363d73)) //10
            mstore(add(ptr, 10), shl(96, implementation_)) //20
            mstore(add(ptr, 30), shl(136, 0x5af43d82803e903d91602b57fd5bf3)) //15
            return(ptr, 45)
        }
    }

    /**
     * @notice calculates salt for a BridgePool implementation contract based on tokenType and version
     * @param tokenType_ type of token (1=ERC20, 2=ERC721)
     * @param version_ version of the implementation
     * @return calculated calculated salt
     */
    function getLocalERCImplementationSalt(uint8 tokenType_, uint16 version_)
        public
        pure
        returns (bytes32)
    {
        string memory tag;
        if (tokenType_ == 1) tag = "LocalERC20";
        else if (tokenType_ == 2) tag = "LocalERC721";
        else if (tokenType_ == 3) tag = "LocalERC1155";
        else tag = "Unknown";
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked(tag)),
                    keccak256(abi.encodePacked(version_))
                )
            );
    }

    /**
     * @dev enables or disables public pool deployment
     **/
    function _togglePublicPoolDeployment() internal {
        _publicPoolDeploymentEnabled = !_publicPoolDeploymentEnabled;
    }

    /**
     * @notice Deploys a new bridge to pass tokens to our chain from the specified ERC contract.
     * The pools are created as thin proxies (EIP1167) routing to versioned implementations identified by correspondent salt.
     * @param tokenType_ type of token (1=ERC20, 2=ERC721)
     * @param ercContract_ address of ERC20 source token contract
     * @param poolVersion_ version of BridgePool implementation to use
     */
    function _deployNewLocalPool(
        uint8 tokenType_,
        address ercContract_,
        uint16 poolVersion_
    ) internal {
        //calculate the unique salt for the bridge pool
        bytes32 bridgePoolSalt = _getBridgePoolSalt(
            ercContract_,
            tokenType_,
            _chainID,
            poolVersion_
        );
        //calculate the address of the pool's logic contract
        address implementation = getMetamorphicContractAddress(
            getLocalERCImplementationSalt(tokenType_, poolVersion_),
            _factoryAddress()
        );
        _implementation = implementation;
        //check if the logic exists for the specified pool
        uint256 implementationSize;
        assembly {
            implementationSize := extcodesize(implementation)
        }
        if (implementationSize == 0) {
            revert BridgePoolFactoryErrors.PoolVersionNotSupported(poolVersion_);
        }
        address contractAddr = _deployStaticPool(bridgePoolSalt);
        IBridgePool(contractAddr).initialize(ercContract_);
        emit BridgePoolCreated(contractAddr, ercContract_);
    }

    /**
     * @notice creates a BridgePool contract with specific salt and bytecode returned by this contract fallback
     * @param salt_ salt of the implementation contract
     * @return contractAddr the address of the BridgePool
     */
    function _deployStaticPool(bytes32 salt_) internal returns (address contractAddr) {
        uint256 contractSize;
        assembly {
            let ptr := mload(0x40)
            mstore(ptr, shl(136, 0x5880818283335afa3d82833e3d82f3))
            contractAddr := create2(0, ptr, 15, salt_)
            contractSize := extcodesize(contractAddr)
        }
        if (contractSize == 0) {
            revert BridgePoolFactoryErrors.StaticPoolDeploymentFailed(salt_);
        }
        return contractAddr;
    }

    /**
     * @notice calculates salt for a BridgePool contract based on ERC contract's address, tokenType, chainID and version_
     * @param tokenContractAddr_ address of ERC contract of BridgePool
     * @param tokenType_ type of token (1=ERC20, 2=ERC721)
     * @param version_ version of the implementation
     * @param chainID_ chain ID
     * @return calculated calculated salt
     */
    function _getBridgePoolSalt(
        address tokenContractAddr_,
        uint8 tokenType_,
        uint256 chainID_,
        uint16 version_
    ) internal pure returns (bytes32) {
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
