// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/errors/BridgePoolFactoryErrors.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/BridgePoolAddressUtil.sol";
import "@openzeppelin/contracts/utils/Strings.sol";

abstract contract BridgePoolFactoryBase is ImmutableFactory {
    //chainid of layer 1 chain, 1 for ether mainnet
    uint256 internal immutable _chainID;
    bool public publicPoolDeploymentEnabled;
    address internal _implementation;
    mapping(bytes => address) internal _logicAddresses;
    //mapping of native and external pools to mapping of pool types to most recent version of logic
    mapping(uint8 => mapping(uint8 => uint16)) internal _logicVersionsDeployed;
    //existing pools
    mapping(address => bool) public poolExists;
    event BridgePoolCreated(
        address poolAddress,
        address ercTokenAddress,
        uint8 poolType,
        uint256 chainID,
        uint16 poolLogicVersion
    );

    modifier onlyFactoryOrPublicEnabled() {
        if (msg.sender != _factoryAddress() && !publicPoolDeploymentEnabled) {
            revert BridgePoolFactoryErrors.PublicPoolDeploymentTemporallyDisabled();
        }
        _;
    }

    constructor() {
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

    /**
     * @notice returns the most recent version of the pool logic
     * @param poolType_ type of pool ie 0 for native, 1 for external
     * @param tokenType_ type of token 0 for ERC20 1 for ERC721 and 2 for ERC1155
     */
    function getLatestPoolLogicVersion(uint8 poolType_, uint8 tokenType_)
        public
        view
        returns (uint16 version)
    {
        version = _logicVersionsDeployed[poolType_][tokenType_];
        if (version == 0)
            revert BridgePoolFactoryErrors.LogicVersionDoesNotExist(poolType_, tokenType_);
    }

    function _deployPoolLogic(
        uint8 poolType_,
        uint8 tokenType_,
        uint16 poolVersion_,
        bytes calldata deployCode_
    ) internal returns (address addr) {
        uint32 codeSize;
        assembly {
            let ptr := mload(0x40)
            calldatacopy(ptr, deployCode_.offset, deployCode_.length)
            // add bytes32 alicenet factory address as parameter to constructor
            addr := create(0, ptr, deployCode_.length)
            codeSize := extcodesize(addr)
        }
        if (codeSize == 0) {
            revert BridgePoolFactoryErrors.FailedToDeployLogic();
        }
        _logicVersionsDeployed[poolType_][tokenType_] += 1;
        //record the depolyed logic address in the mapping
        _logicAddresses[_getImplementationAddressKey(poolType_, tokenType_, poolVersion_)] = addr;
    }

    /**
     * @dev enables or disables public pool deployment
     **/
    function _togglePublicPoolDeployment() internal {
        publicPoolDeploymentEnabled = !publicPoolDeploymentEnabled;
    }

    function _deployNewPool(
        uint8 poolType_,
        uint8 tokenType_,
        address ercContract_,
        uint16 poolVersion_,
        uint256 chainID_,
        bytes calldata initCallData
    ) internal {
        // get the unique salt for the bridge pool
        bytes32 bridgePoolSalt = BridgePoolAddressUtil.getBridgePoolSalt(
            ercContract_,
            tokenType_,
            chainID_,
            poolVersion_
        );
        //look up the address in the _logicAddresses mapping
        address implementation = _logicAddresses[
            _getImplementationAddressKey(poolType_, tokenType_, poolVersion_)
        ];
        //check if the logic exists for the specified pool
        uint256 implementationSize;
        assembly {
            implementationSize := extcodesize(implementation)
        }
        if (implementationSize == 0 || implementation == address(0)) {
            revert BridgePoolFactoryErrors.PoolLogicNotSupported();
        }
        address contractAddr = _deployStaticPool(bridgePoolSalt);
        _initializeContract(contractAddr, initCallData);
        IBridgePool(contractAddr).initialize(ercContract_);
        emit BridgePoolCreated(contractAddr, ercContract_, poolType_, chainID_, poolVersion_);
    }

    function _initializeContract(address contract_, bytes calldata initCallData_) internal {
        assembly ("memory-safe") {
            if iszero(iszero(initCallData_.length)) {
                let ptr := mload(0x40)
                mstore(0x40, add(initCallData_.length, ptr))
                calldatacopy(ptr, initCallData_.offset, initCallData_.length)
                if iszero(call(gas(), contract_, 0, ptr, initCallData_.length, 0x00, 0x00)) {
                    ptr := mload(0x40)
                    mstore(0x40, add(returndatasize(), ptr))
                    returndatacopy(ptr, 0x00, returndatasize())
                    revert(ptr, returndatasize())
                }
            }
        }
    }

    /**
     * @notice deploys a bridge pool clone given a salt, the implementation slot must be set with the correct implementation contract address
     * @param salt_ use BridgePoolAddressUtil.getBridgePoolSaltto generate the unique salt for the specified pool
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
        if (contractSize == 0 || contractAddr == address(0)) {
            revert BridgePoolFactoryErrors.StaticPoolDeploymentFailed(salt_);
        }
        poolExists[contractAddr] = true;
        return contractAddr;
    }

    /**
     * @notice calculates salt for a BridgePool implementation contract based on tokenType and version
     * @param poolType_ type of pool (0=Native, 1=External)
     * @param tokenType_ type of token (0=ERC20, 1=ERC721, 2=ERC1155)_
     * @param version_ version of bridge pool logic
     * @return key unique key used to reference implementation address from
     */
    function _getImplementationAddressKey(
        uint8 poolType_,
        uint8 tokenType_,
        uint16 version_
    ) internal pure returns (bytes memory key) {
        key = abi.encode(version_, poolType_, tokenType_);
    }
}
