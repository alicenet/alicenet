// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/AliceNetFactory.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/interfaces/IBridgePool.sol";
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
    event Deposited(
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

    // function getStorageCode() external view returns (bytes memory) {}

    /**
     * @dev this function can only be called by the btoken contract, when called this function will
     * calculate the target pool address with the information in call data.
     * The fee will be passed in as a parameter by the user and btoken verify the deposit happened before
     * forwarding the call to this function.
     * @param data contains information needed to perform deposit, ERCCONTRACTADDRESS, ChainID, Version
     */
    function routeDeposit(bytes memory data) public onlyBToken returns (uint256 btokenFeeAmount) {
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
        BridgePool(poolAddress).deposit(depositCallData.depositAmount);
        //get the fee to deposit a token into the bridge
        btokenFeeAmount = 10;
        emit Deposited(
            nonce,
            depositCallData.ERCContract,
            msg.sender,
            depositCallData.tokenType,
            depositCallData.number
        );
        nonce++;
    }

    /**
     @dev this function checks if the pool exists
    */
    function poolExists(address Pool) public view returns (bool) {
        if (extcodesize(pool) == 0) {
            return false;
        }
        return true;
    }

    /**
     * @notice deployNewPool
     * @param erc20Contract_ address of ERC20 token contract
     */
    // function deployNewLocalPool(
    //     address erc20Contract_,
    //     uint16 implementationVersion_
    // ) public {
    //     bytes32 bridgePoolSalt = getLocalBridgePoolSalt(erc20Contract_);
    //     _implementation = getMetamorphicContractAddress(
    //         getImplementationSalt(implementationVersion_),
    //         _factoryAddress()
    //     );
    //     address contractAddr = _deployStaticPool(bridgePoolSalt, initCallData);
    //     emit BridgePoolCreated(contractAddr);
    // }

    /**
     * @notice getSaltFromAddress calculates salt for a BridgePool contract based on ERC20 contract's address
     * @param version_ address of ERC20 contract of BridgePool
     * @return calculated salt
     */
    function getLocalERC20ImplementationSalt(uint16 version_) public pure returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked("LocalERC20")),
                    keccak256(abi.encodePacked(version_))
                )
            );
    }

    function getLocalERC721ImplementationSalt(uint16 version_) public pure returns (bytes32) {
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked("LocalERC721")),
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

    function codeCopy(address addr) public view returns (bytes memory) {
        assembly {
            let ptr := mload(0x40)
            extcodecopy(addr, ptr, 0, extcodesize(addr))
            return(ptr, extcodesize(addr))
        }
    }

    function _deployStaticPool(bytes32 salt_, bytes memory initCallData_)
        internal
        returns (address contractAddr)
    {
        uint256 response;
        uint256 contractsize;
        assembly {
            let ptr := mload(0x40)
            mstore(ptr, shl(144, 0x5880818283335afa3d82833e3d82f3))
            //mstore(ptr, shl(136, 0x5880818283335afa3d82833e3d91f3))
            contractAddr := create2(0, ptr, 15, salt_)
            response := returndatasize()
            contractsize := extcodesize(contractAddr)
            //if the returndatasize is not 0 revert with the error message
            if iszero(iszero(returndatasize())) {
                returndatacopy(0x00, 0x00, returndatasize())
                revert(0, returndatasize())
            }
            //if contractAddr or code size at contractAddr is 0 revert with deploy fail message
            if or(iszero(contractAddr), iszero(extcodesize(contractAddr))) {
                mstore(0, "static pool deploy failed")
                revert(0, 0x20)
            }
        }
        console.log("bytecode");
        console.logBytes(codeCopy(contractAddr));
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
