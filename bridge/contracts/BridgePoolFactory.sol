// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/AliceNetFactory.sol";
import "contracts/BridgePoolV1.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

/// @custom:salt BridgePoolFactory
/// @custom:deploy-type deployUpgradeable
contract BridgePoolFactory is
    Initializable,
    ImmutableFactory,
    ImmutableBridgePoolFactory,
    ImmutableBridgePoolDepositNotifier
{
    uint256 internal immutable _networkId;
    string internal constant _BRIDGE_POOL_TAG = "ERC";
    event BridgePoolCreated(address contractAddr);
    /** 
    @dev slot for storing implementation address
    */
    address private _implementation;

    constructor(uint256 networkId_) ImmutableFactory(msg.sender) {
        _networkId = networkId_;
    }

    function initialize(address implementation_) public initializer {
        _implementation = implementation_;
    }

    /**
     * @notice deployNewPool delegates call to this contract's method "deployViaFactoryLogic" through alicenet factory
     * @param erc20Contract_ address of ERC20 token contract
     * @param token address of bridge token contract
     */
    function deployNewPool(address erc20Contract_, address token) public {
        bytes memory initCallData = abi.encodePacked(erc20Contract_, token);
        bytes32 salt = getSaltFromERC20Address(erc20Contract_);
        address contractAddr = _deployStaticPool(salt, initCallData);
        BridgePoolV1(payable(contractAddr)).initialize(erc20Contract_, token);
        emit BridgePoolCreated(contractAddr);
    }

    /**
     * @notice getSaltFromAddress calculates salt for a BridgePool contract based on ERC20 contract's address
     * @param erc20Contract_ address of ERC20 contract of BridgePool
     * @return calculated salt
     */
    function getSaltFromERC20Address(address erc20Contract_)
        public
        view
        returns (
            //onlyBridgePoolDepositNotifier
            bytes32
        )
    {
        return
            keccak256(
                bytes.concat(
                    keccak256(abi.encodePacked(erc20Contract_)),
                    keccak256(abi.encodePacked(_BRIDGE_POOL_TAG)),
                    keccak256(abi.encodePacked(_networkId))
                )
            );
    }

    //salt:keccak(keccak( keccak(ERCADDRESS),keccak(chainID), keccak(Version))

    function getStaticPoolContractAddress(bytes32 _salt, address _factory)
        public
        pure
        returns (address)
    {
        // byte code for metamorphic contract
        // 58808182335afa3d82833e3d91f3 does not work
        // bytes32 metamorphicContractBytecodeHash_ = 0xcd77112ba3315c30f6863dae90cb281bf2f644ef3fd9d21e53d1968182daa472;
        // 58808182335afa3d36363e3d36f3 works
        bytes32 metamorphicContractBytecodeHash_ = 0x031959e1c8af62b96e38494efad5389b26bb3eb10374abdcf4e66994dffd7bd5;
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
        uint256 response;
        uint256 contractsize;
        assembly {
            let ptr := mload(0x40)
            //   mstore(ptr, shl(144, 0x58808182335afa3d36363e3d36f3))
            mstore(ptr, shl(136, 0x5880818283335afa3d82833e3d91f3))
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
