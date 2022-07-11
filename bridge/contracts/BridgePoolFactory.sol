// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/AliceNetFactory.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePool.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "hardhat/console.sol";
import "contracts/libraries/errorCodes/BridgePoolFactoryErrorCodes.sol";

/// @custom:salt BridgePoolFactory
/// @custom:deploy-type deployUpgradeable
contract BridgePoolFactory is
    Initializable,
    ImmutableFactory,
    ImmutableBridgePoolFactory,
    ImmutableBridgePoolDepositNotifier
{
    uint256 internal immutable _networkId;
    event BridgePoolCreated(address contractAddr);
    address private _implementation;

    constructor(uint256 networkId_) ImmutableFactory(msg.sender) {
        _networkId = networkId_;
    }

    /**
     * @notice deployNewPool
     * @param ERContract_ address of ERC20 token contract
     * @param implementationVersion_ version of the BridgePool Implementation
     */
    function deployNewPool(address ERContract_, uint16 implementationVersion_) public {
        bytes memory initCallData = abi.encodeWithSignature("initialize(address)", ERContract_);
        _implementation = getMetamorphicContractAddress(
            getImplementationSalt(implementationVersion_),
            _factoryAddress()
        );
        uint256 implementationSize;
        address implementation = _implementation;
        assembly {
            implementationSize := extcodesize(implementation)
        }
        require(
            implementationSize > 0,
            string(
                abi.encodePacked(
                    BridgePoolFactoryErrorCodes
                        .BRIDGEPOOLFACTORY_UNEXISTENT_BRIDGEPOOL_IMPLEMENTATION_VERSION
                )
            )
        );
        bytes32 bridgePoolSalt = getLocalBridgePoolSalt(ERContract_);
        address contractAddr = _deployStaticPool(bridgePoolSalt, initCallData);
        IBridgePool(contractAddr).initialize(ERContract_);
        emit BridgePoolCreated(contractAddr);
    }

    /**
     * @notice getSaltFromAddress calculates salt for a BridgePool contract based on ERC20 contract's address
     * @param version_ address of ERC20 contract of BridgePool
     * @return calculated salt
     */
    function getImplementationSalt(uint16 version_) public view returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked("LocalERC20")),
                    keccak256(abi.encodePacked(version_))
                )
            );
    }

    /**
     * @notice getSaltFromAddress calculates salt for a BridgePool contract based on ERC contract's address
     * @param ERCContract_ address of ERC contract of BridgePool
     * @return calculated salt
     */
    function getLocalBridgePoolSalt(address ERCContract_) public view returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked(ERCContract_)),
                    keccak256(abi.encodePacked(_networkId))
                )
            );
    }

    function getStaticPoolContractAddress(bytes32 _salt, address _factory)
        public
        pure
        returns (address)
    {
        //0x5880818283335afa3d82833e3d82f3
        bytes32 metamorphicContractBytecodeHash_ = 0xf231e946a2f88d89eafa7b43271c54f58277304b93ac77d138d9b0bb3a989b6d;
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

    function _deployStaticPool(bytes32 salt_, bytes memory initCallData_)
        internal
        returns (address contractAddr)
    {
        address contractAddr;
        uint256 contractSize;

        assembly {
            let ptr := mload(0x40)
            mstore(ptr, shl(136, 0x5880818283335afa3d82833e3d82f3))
            contractAddr := create2(0, ptr, 15, salt_)
            contractSize := extcodesize(contractAddr)
        }
        require(
            contractSize > 0,
            string(
                abi.encodePacked(
                    BridgePoolFactoryErrorCodes.BRIDGEPOOLFACTORY_UNABLE_TO_DEPLOY_BRIDGEPOOL
                )
            )
        );
        // if (initCallData_.length > 0) {
        //     AliceNetFactory(_factoryAddress()).initializeContract(contractAddr, initCallData_);
        // }
        return contractAddr;
    }

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
}
