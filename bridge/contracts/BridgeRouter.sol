// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/AliceNetFactory.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePool.sol";
import "contracts/libraries/errorCodes/BridgeRouterErrorCodes.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "hardhat/console.sol";

/// @custom:salt Fees
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

    constructor(uint256 networkId_) ImmutableFactory(msg.sender) {
        _networkId = networkId_;
    }

    /**
     * @dev this function can only be called by the btoken contract, when called this function will
     * calculate the target pool address with the information in call data.
     * The fee will be passed in as a parameter by the user and btoken verify the deposit happened before
     * forwarding the call to this function.
     * @param data contains information needed to perform deposit, ERCCONTRACTADDRESS, ChainID, Version
     */
    function routeDeposit(
        address msgSender,
        uint256 maxTokens,
        bytes memory data
    ) public onlyBToken returns (uint256 btokenFeeAmount) {
        //get the fee to deposit a token into the bridge
        btokenFeeAmount = 10;
        require(maxTokens >= btokenFeeAmount, "BridgeRouter: ERROR insufficient funds");
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
        IBridgePool(poolAddress).deposit(1, msgSender, depositCallData.number);

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
     * @notice deployNewLocalPool
     * @param ercContract_ address of ERC20 token contract
     */
    function deployNewLocalPool(
        uint8 type_,
        address ercContract_,
        uint16 implementationVersion_
    ) public {
        bytes memory initCallData = abi.encodeWithSignature("inititalize(address)", ercContract_);
        bytes32 bridgePoolSalt = getBridgePoolSalt(
            ercContract_,
            type_,
            _networkId,
            implementationVersion_
        );
        _implementation = getMetamorphicContractAddress(
            getLocalERCImplementationSalt(type_, implementationVersion_),
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
        address contractAddr = _deployStaticPool(bridgePoolSalt, initCallData);
        IBridgePool(contractAddr).initialize(ercContract_);
        emit BridgePoolCreated(contractAddr);
    }

    /**
     * @notice getLocalERCImplementationSalt calculates salt for a BridgePool Implementation contract based on tokenType and version
     * @param version_ address of ERC20 contract of BridgePool
     * @return calculated salt
     */
    function getLocalERCImplementationSalt(uint8 tokenType, uint16 version_)
        public
        pure
        returns (bytes32)
    {
        string memory tag;
        if (tokenType == 1) tag = "LocalERC20";
        if (tokenType == 2) tag = "LocalERC721";
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked(tag)),
                    keccak256(abi.encodePacked(version_))
                )
            );
    }

    /**
     * @notice getSaltFromAddress calculates salt for a BridgePool contract based on ERC20 contract's address
     * @param tokenContractAddr_ address of ERC20 contract of BridgePool
     * @return calculated salt
     */
    function getBridgePoolSalt(
        address tokenContractAddr_,
        uint8 type_,
        uint256 chainID_,
        uint16 version_
    ) public pure returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked(tokenContractAddr_)),
                    keccak256(abi.encodePacked(type_)),
                    keccak256(abi.encodePacked(chainID_)),
                    keccak256(abi.encodePacked(version_))
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
                abi.encodePacked(BridgeRouterErrorCodes.BRIDGEROUTER_UNABLE_TO_DEPLOY_BRIDGEPOOL)
            )
        );
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
