// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/errors/BridgePoolFactoryErrors.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/BridgePoolAddressUtil.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

abstract contract BridgePoolFactoryBase is ImmutableFactory {
    enum TokenType {
        ERC20,
        ERC721,
        ERC1155
    }
    enum PoolType {
        NATIVE,
        EXTERNAL
    }
    //chainid of layer 1 chain, 1 for ether mainnet
    uint256 internal immutable _chainID;
    bool public publicPoolDeploymentEnabled;
    address internal _implementation;
    mapping(string => address) internal _logicAddresses;
    //mapping of native and external pools to mapping of pool types to most recent version of logic
    mapping(PoolType => mapping(TokenType => uint16)) logicVersionsDeployed_;
    event BridgePoolCreated(address poolAddress, address token);

    modifier onlyFactoryOrPublicPoolDeploymentEnabled() {
        if (msg.sender != _factoryAddress() && !publicPoolDeploymentEnabled) {
            revert BridgePoolFactoryErrors.PublicPoolDeploymentTemporallyDisabled();
        }
        _;
    }

    constructor() ImmutableFactory(msg.sender) {
        _chainID = block.chainid;
    }

    // NativeERC20V!
    /**
     * @notice returns bytecode for a Minimal Proxy (EIP-1167) that routes to BridgePool implementation
     */
    // solhint-disable-next-line
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

    function _deployPoolLogic(
        uint8 tokenType_,
        uint256 chainId_,
        uint256 value_,
        bytes calldata deployCode_
    ) internal returns (address) {
        address addr;
        uint32 codeSize;
        bool native = true;
        uint16 version;
        assembly {
            let ptr := mload(0x40)
            calldatacopy(ptr, deployCode_.offset, deployCode_.length)
            addr := create(value_, ptr, deployCode_.length)
            codeSize := extcodesize(addr)
        }
        if (codeSize == 0) {
            revert BridgePoolFactoryErrors.FailedToDeployLogic();
        }
        if (chainId_ != _chainID) {
            native = false;
            version = logicVersionsDeployed_[PoolType.EXTERNAL][TokenType(tokenType_)] + 1;
            logicVersionsDeployed_[PoolType.EXTERNAL][TokenType(tokenType_)] = version;
        } else {
            version = logicVersionsDeployed_[PoolType.NATIVE][TokenType(tokenType_)] + 1;
            logicVersionsDeployed_[PoolType.NATIVE][TokenType(tokenType_)] = version;
        }
        _logicAddresses[_getImplementationAddressKey(tokenType_, version, native)] = addr;
        return addr;
    }

    /**
     * @dev enables or disables public pool deployment
     **/
    function _togglePublicPoolDeployment() internal {
        publicPoolDeploymentEnabled = !publicPoolDeploymentEnabled;
    }

    /**
     * @notice Deploys a new bridge to pass tokens to layer 2 chain from the specified ERC contract.
     * The pools are created as thin proxies (EIP1167) routing to versioned implementations identified by correspondent salt.
     * @param tokenType_ type of token (0=ERC20, 1=ERC721, 2=ERC1155)
     * @param ercContract_ address of ERC20 source token contract
     * @param poolVersion_ version of BridgePool implementation to use
     */
    function _deployNewNativePool(
        uint8 tokenType_,
        address ercContract_,
        uint16 poolVersion_
    ) internal {
        bool native = true;
        //calculate the unique salt for the bridge pool
        bytes32 bridgePoolSalt = BridgePoolAddressUtil.getBridgePoolSalt(
            ercContract_,
            tokenType_,
            _chainID,
            poolVersion_
        );
        //calculate the address of the pool's logic contract
        address implementation = _logicAddresses[
            _getImplementationAddressKey(tokenType_, poolVersion_, native)
        ];
        _implementation = implementation;
        //check if the logic exists for the specified pool
        //TODO determin if this step is still necessary
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

    function getLatestPoolLogicVersion() public view returns(uint16) {

    }

    /**
     * @notice calculates salt for a BridgePool implementation contract based on tokenType and version
     * @param tokenType_ type of token (0=ERC20, 1=ERC721, 2=ERC1155)
     * @param version_ version of the implementation
     * @param native_ boolean flag to specifier native or external token pools
     * @return calculated key
     */
    function _getImplementationAddressKey(
        uint8 tokenType_,
        uint16 version_,
        bool native_
    ) internal pure returns (string memory) {
        string memory key;
        if (native_) {
            key = "Native";
        } else {
            key = "External";
        }
        if (tokenType_ == uint8(TokenType.ERC20)) {
            key = string.concat(key, "ERC20");
        } else if (tokenType_ == uint8(TokenType.ERC721)) {
            key = string.concat(key, "ERC721");
        } else if (tokenType_ == uint8(TokenType.ERC1155)) {
            key = string.concat(key, "ERC1155");
        }
        key = string.concat(key, "V", Strings.toString(version_));
        return key;
    }
}
