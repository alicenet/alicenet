// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/AliceNetFactory.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/libraries/errorCodes/BridgeRouterErrorCodes.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errorCodes/CircuitBreakerErrorCodes.sol";
import "hardhat/console.sol";

/// @custom:salt BridgeRouter
/// @custom:deploy-type deployUpgradeable
contract BridgeRouter is
    Initializable,
    ImmutableFactory,
    ImmutableBridgeRouter,
    ImmutableBridgePoolDepositNotifier,
    ImmutableBToken
{
    event DepositedERCToken(
        uint256 nonce,
        address ercContract,
        address owner,
        uint8 ERCType,
        uint256 number, // If fungible, this is the amount. If non-fungible, this is the id
        uint256 chainID
    );

    struct DepositCallData {
        address ERCContract;
        uint8 tokenType;
        uint256 number;
        uint256 chainID;
        uint16 poolVersion;
    }

    uint256 internal immutable _networkId;
    event BridgePoolCreated(address contractAddr);
    address private _implementation;
    uint256 nonce;
    bool private publicPoolDeploymentEnabled = false;

    modifier onlyFactoryOrPublicPoolDeploymentEnabled() {
        require(
            msg.sender == _factoryAddress() || publicPoolDeploymentEnabled == true,
            string(
                abi.encodePacked(
                    BridgeRouterErrorCodes.BRIDGEROUTER_POOL_DEPLOYMENT_TEMPORALLY_DISABLED
                )
            )
        );
        _;
    }

    constructor(uint256 networkId_) ImmutableFactory(msg.sender) {
        _networkId = networkId_;
    }

    /**
     * @dev enables or disables public pool deployment
     **/
    function togglePublicPoolDeployment() public onlyFactory {
        if (publicPoolDeploymentEnabled == true) publicPoolDeploymentEnabled = false;
        else publicPoolDeploymentEnabled = true;
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
        bytes32 bridgePoolSalt = getBridgePoolSalt(
            ercContract_,
            tokenType_,
            _networkId,
            implementationVersion_
        );
        _implementation = getMetamorphicContractAddress(
            getLocalERCImplementationSalt(tokenType_, implementationVersion_),
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
                    BridgeRouterErrorCodes.BRIDGEROUTER_UNEXISTENT_BRIDGEPOOL_IMPLEMENTATION_VERSION
                )
            )
        );
        address contractAddr = _deployStaticPool(bridgePoolSalt);
        IBridgePool(contractAddr).initialize(ercContract_);
        emit BridgePoolCreated(contractAddr);
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
        if (tokenType_ == 1) {
            tag = "LocalERC20";
        } else if (tokenType_ == 2) tag = "LocalERC721";
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
     * @notice calculates salt for a BridgePool contract based on ERC contract's address, tokenType, chainID and version_
     * @param tokenContractAddr_ address of ERC contract of BridgePool
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
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked(tokenContractAddr_)),
                    keccak256(abi.encodePacked(tokenType_)),
                    keccak256(abi.encodePacked(chainID_)),
                    keccak256(abi.encodePacked(version_))
                )
            );
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
        require(
            contractSize > 0,
            string(
                abi.encodePacked(BridgeRouterErrorCodes.BRIDGEROUTER_UNABLE_TO_DEPLOY_BRIDGEPOOL)
            )
        );
        return contractAddr;
    }

    /**
     * @dev this function can only be called by the btoken contract, when called this function will
     * calculate the target pool address with the information in call data.
     * The fee will be passed in as a parameter by the user and btoken verify the deposit happened before
     * forwarding the call to this function.
     * @param msgSender the original sender of the value
     * @param maxTokens max number of bTokens willing to pay for the deposit
     * @param data contains information needed to perform deposit, ERCCONTRACTADDRESS, ChainID, Version
     */
    function routeDeposit(
        address msgSender,
        uint256 maxTokens,
        bytes memory data
    ) public onlyBToken returns (uint256 btokenFeeAmount) {
        //get the fee to deposit a token into the bridge
        btokenFeeAmount = 10;
        require(
            maxTokens >= btokenFeeAmount,
            string(abi.encodePacked(BridgeRouterErrorCodes.BRIDGEROUTER_INSUFFICIENT_FUNDS))
        );
        // use abi decode to extract the information out of data
        DepositCallData memory depositCallData = abi.decode(data, (DepositCallData));
        //encode the salt with the information from
        bytes32 poolSalt = getBridgePoolSalt(
            depositCallData.ERCContract,
            depositCallData.tokenType,
            depositCallData.chainID,
            depositCallData.poolVersion
        );
        // calculate the address of the pool
        address poolAddress = getStaticPoolContractAddress(poolSalt, address(this));

        //call the pool to initiate deposit
        IBridgePool(poolAddress).deposit(msgSender, depositCallData.number);

        emit DepositedERCToken(
            nonce,
            depositCallData.ERCContract,
            msg.sender,
            depositCallData.tokenType,
            depositCallData.number,
            _networkId
        );
        nonce++;
        return btokenFeeAmount;
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
}
