// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;


import "contracts/libraries/errors/BridgePoolFactoryErrors.sol";
import "contracts/libraries/factory/BridgePoolFactoryBase.sol";
/// @custom:salt BridgePoolFactory
/// @custom:deploy-type deployUpgradeable
contract BridgePoolFactory is BridgePoolFactoryBase {

    constructor(uint256 chainID_) BridgePoolFactoryBase(chainID_){

    }

     /**
     * @dev enables or disables public pool deployment
     **/
    function togglePublicPoolDeployment() public onlyFactory {
        _togglePublicPoolDeployment();
    }

     /**
     * @notice Deploys a new bridge to pass tokens to our chain from the specified ERC contract.
     * The pools are created as thin proxies (EIP1167) routing to versioned implementations identified by correspondent salt.
     * @param tokenType_ type of token (1=ERC20, 2=ERC721)
     * @param ercContract_ address of ERC20 source token contract
     * @param implementationVersion_ version of BridgePool implementation to use
     */
    function deployNewLocalPool(
        uint8 tokenType_,
        address ercContract_,
        uint16 implementationVersion_
    ) public onlyFactoryOrPublicPoolDeploymentEnabled {
        _deployNewLocalPool(tokenType_, ercContract_, implementationVersion_);
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
        return _getBridgePoolSalt(
            tokenContractAddr_,
            tokenType_,
            chainID_,
            version_
        );
    }

}