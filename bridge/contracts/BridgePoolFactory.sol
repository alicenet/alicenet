// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/libraries/errors/BridgePoolFactoryErrors.sol";
import "contracts/libraries/factory/BridgePoolFactoryBase.sol";
import "contracts/utils/auth/ImmutableETHDKG.sol";
import "hardhat/console.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

/// @custom:salt BridgePoolFactory
/// @custom:deploy-type deployUpgradeable
contract BridgePoolFactory is
    ImmutableFactory,
    ImmutableSnapshots,
    ImmutableETHDKG,
    BridgePoolFactoryBase,
    Initializable
{
    address internal _aliceNetRegistry;
    address internal _ethDKGAddress;

    constructor()
        ImmutableFactory(msg.sender)
        ImmutableSnapshots()
        ImmutableETHDKG()
        BridgePoolFactoryBase()
    {}

    function initialize() public onlyFactory initializer {
        _aliceNetRegistry = msg.sender;
    }

    /**
     * @notice Deploys a new bridge to pass ERC tokens to external chains.
     * The pools are created as thin proxies (EIP1167) routing to versioned implementations identified by corresponding salt.
     * @param tokenType_ type of token (0=ERC20, 1=ERC721, 2=ERC1155)
     * @param ercContract_ address of ERC20 source token contract
     * @param poolVersion_ version of BridgePool implementation to use
     */
    function deployNewExternalPool(
        uint8 tokenType_,
        address ercContract_,
        uint16 poolVersion_,
        uint256 chainID_,
        bytes calldata initCallData
    ) public onlyFactoryOrPublicEnabled {
        _deployNewPool(1, tokenType_, ercContract_, poolVersion_, chainID_, initCallData);
    }

    /**
     * @notice Deploys a new bridge to pass tokens to layer 2 chain from the specified ERC contract.
     * The pools are created as thin proxies (EIP1167) routing to versioned implementations identified by correspondent salt.
     * @param tokenType_ type of token (0=ERC20, 1=ERC721, 2=ERC1155)
     * @param ercContract_ address of ERC20 source token contract
     * @param poolVersion_ version of BridgePool implementation to use
     */
    function deployNewNativePool(
        uint8 tokenType_,
        address ercContract_,
        uint16 poolVersion_,
        uint256 chainID_,
        bytes calldata initCallData
    ) public onlyFactoryOrPublicEnabled {
        _deployNewPool(0, tokenType_, ercContract_, poolVersion_, chainID_, initCallData);
    }

    /**
     * @notice allows factory to deploy arbitrary types of pools
     */
    function deployNewArbitraryPool(
        uint8 poolType_,
        uint8 tokenType_,
        address ercContract_,
        uint16 poolVersion_,
        uint256 chainID_,
        bytes calldata initCallData
    ) public onlyFactory {
        _deployNewPool(poolType_, tokenType_, ercContract_, poolVersion_, chainID_, initCallData);
    }

    /**
     * @notice deploys logic for bridge pools and stores it in a logicAddresses mapping
     * @param poolType_ type of pool (0=native, 1=external)
     * @param tokenType_ type of token (1=ERC20, 2=ERC721)
     * @param poolVersion_ version of bridge pool logic to deploy
     * @param deployCode_ logic contract deployment bytecode
     */
    function deployPoolLogic(
        uint8 poolType_,
        uint8 tokenType_,
        uint16 poolVersion_,
        bytes calldata deployCode_
    ) public onlyFactory returns (address) {
        return _deployPoolLogic(poolType_, tokenType_, poolVersion_, deployCode_);
    }

    /**
     * @dev enables or disables public pool deployment
     **/
    function togglePublicPoolDeployment() public onlyFactory {
        _togglePublicPoolDeployment();
    }

    /**
     *
     * @notice this is intended for use as a failsafe incase we want
     * to deploy a standalone registry contract in the future
     * @param newAddress_ address of the new registry contract
     *
     */
    function changeAliceNetRegistryAddress(address newAddress_) public onlyFactory {
        _aliceNetRegistry = newAddress_;
    }

    /**
     * @notice calculates bridge pool address with associated bytes32 salt
     * @param bridgePoolSalt_ bytes32 salt associated with the pool, calculated with getBridgePoolSalt
     * @return poolAddress calculated calculated bridgePool Address
     */
    function lookupBridgePoolAddress(
        bytes32 bridgePoolSalt_
    ) public view returns (address poolAddress) {
        poolAddress = BridgePoolAddressUtil.getBridgePoolAddress(bridgePoolSalt_, address(this));
    }

    /**
     * @notice returns alicenet factory address, initializing
     * immutable address in pool logic
     * @dev use this to get factory address and lookup other
     * contracts from the constructor, when deploying pool logic
     */
    function getRegistryAddress() public view returns (address registryAddress) {
        registryAddress = _aliceNetRegistry;
    }

    /**
     * @notice calculates salt for a BridgePool contract based on ERC contract's address, tokenType, chainID and version_
     * @param tokenContractAddr_ address of ERC Token contract
     * @param tokenType_ type of token (1=ERC20, 2=ERC721)
     * @param version_ version of the implementation
     * @param chainID_ chain ID
     * @return calculated calculated salt
     */
    function getBridgePoolSalt(
        address tokenContractAddr_,
        uint8 tokenType_,
        uint256 chainID_,
        uint16 version_
    ) public pure returns (bytes32) {
        return
            BridgePoolAddressUtil.getBridgePoolSalt(
                tokenContractAddr_,
                tokenType_,
                chainID_,
                version_
            );
    }
}
