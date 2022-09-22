// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/libraries/errors/BridgePoolFactoryErrors.sol";
import "contracts/libraries/factory/BridgePoolFactoryBase.sol";

/// @custom:salt BridgePoolFactory
/// @custom:deploy-type deployUpgradeable
contract BridgePoolFactory is BridgePoolFactoryBase {
    constructor(uint256 chainID_) BridgePoolFactoryBase(chainID_) {}

    
    /**
     * @notice Deploys a new bridge to pass tokens to our chain from the specified ERC contract.
     * The pools are created as thin proxies (EIP1167) routing to versioned implementations identified by corresponding salt.
     * @param tokenType_ type of token (1=ERC20, 2=ERC721)
     * @param ercContract_ address of ERC20 source token contract
     * @param implementationVersion_ version of BridgePool implementation to use
     */
    function deployNewNativePool(
        uint8 tokenType_,
        address ercContract_,
        uint16 implementationVersion_
    ) public onlyFactoryOrPublicPoolDeploymentEnabled {
        _deployNewNativePool(tokenType_, ercContract_, implementationVersion_);
    }
    
     function deployPoolLogic(uint8 tokenType_, uint256 chainId_, uint16 version_, uint256 value_, bytes calldata deployCode_) public onlyFactory returns(address){
        return _deployPoolLogic(tokenType_, chainId_, version_, value_, deployCode_);
    }

    /**
     * @dev enables or disables public pool deployment
     **/
    function togglePublicPoolDeployment() public onlyFactory {
        _togglePublicPoolDeployment();
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
        return BridgePoolAddressUtil._getBridgePoolSalt(tokenContractAddr_, tokenType_, chainID_, version_);
    }
}
