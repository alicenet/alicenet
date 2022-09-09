// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/Proxy.sol";
import "contracts/utils/DeterministicAddress.sol";
import "contracts/utils/ArbitraryDeterministicAddress.sol";
import "contracts/libraries/proxy/ProxyUpgrader.sol";
import "contracts/interfaces/IProxy.sol";
import "contracts/libraries/errors/AliceNetFactoryBaseErrors.sol";
import "contracts/AToken.sol";

abstract contract AliceNetFactoryBase is
    DeterministicAddress,
    ArbitraryDeterministicAddress,
    ProxyUpgrader
{
    struct MultiCallArgs {
        address target;
        uint256 value;
        bytes data;
    }

    struct ContractInfo {
        bool exist;
        address logicAddr;
    }

    /**
    @dev owner role for privileged access to functions
    */
    address private _owner;

    /**
    @dev array to store list of contract salts
    */
    bytes32[] private _contracts;

    /**
    @dev slot for storing implementation address
    */
    address private _implementation;

    address private immutable _proxyTemplate;

    bytes8 private constant _UNIVERSAL_DEPLOY_CODE = 0x38585839386009f3;

    mapping(bytes32 => ContractInfo) internal _externalContractRegistry;

    address internal immutable _aTokenAddress;
    bytes32 internal immutable _aTokenCreationCodeHash;
    bytes32 internal constant _aTokenSalt =
        0x41546f6b656e0000000000000000000000000000000000000000000000000000;

    /**
     *@dev events that notify of contract deployment
     */
    event Deployed(bytes32 salt, address contractAddr);
    event DeployedTemplate(address contractAddr);
    event DeployedStatic(address contractAddr);
    event DeployedRaw(address contractAddr);
    event DeployedProxy(address contractAddr);

    // modifier restricts caller to owner or self via multicall
    modifier onlyOwner() {
        _requireAuth(msg.sender == address(this) || msg.sender == owner());
        _;
    }

    /**
     * @dev The constructor encodes the proxy deploy byte code with the _UNIVERSAL_DEPLOY_CODE at the
     * head and the factory address at the tail, and deploys the proxy byte code using create OpCode.
     * The result of this deployment will be a contract with the proxy contract deployment bytecode with
     * its constructor at the head, runtime code in the body and constructor args at the tail. The
     * constructor then sets proxyTemplate_ state var to the deployed proxy template address the deploy
     * account will be set as the first owner of the factory.
     */
    constructor(address legacyToken_) {
        bytes memory proxyDeployCode = abi.encodePacked(
            //8 byte code copy constructor code
            _UNIVERSAL_DEPLOY_CODE,
            type(Proxy).creationCode,
            bytes32(uint256(uint160(address(this))))
        );
        //variable to store the address created from create(the location of the proxy template contract)
        address addr;
        assembly {
            //deploys the proxy template contract
            addr := create(0, add(proxyDeployCode, 0x20), mload(proxyDeployCode))
            if iszero(addr) {
                //if contract creation fails, we want to return any err messages
                returndatacopy(0x00, 0x00, returndatasize())
                //revert and return errors
                revert(0x00, returndatasize())
            }
        }
        //State var that stores the proxyTemplate address
        _proxyTemplate = addr;
        //State var that stores the _owner address
        _owner = msg.sender;

        // Deploying ALCA
        bytes memory aTokenCreationCode = abi.encodePacked(
            type(AToken).creationCode,
            bytes32(uint256(uint160(legacyToken_)))
        );
        address aTokenAddress;
        assembly {
            aTokenAddress := create2(
                0,
                add(aTokenCreationCode, 0x20),
                mload(aTokenCreationCode),
                _aTokenSalt
            )
        }
        _codeSizeZeroRevert((_extCodeSize(aTokenAddress) != 0));
        _aTokenAddress = aTokenAddress;
        _aTokenCreationCodeHash = keccak256(abi.encodePacked(aTokenCreationCode));
    }

    // solhint-disable payable-fallback
    /**
     * @dev fallback function returns the address of the most recent deployment of a template
     */
    fallback() external {
        assembly {
            mstore(returndatasize(), sload(_implementation.slot))
            return(returndatasize(), 0x20)
        }
    }

    /**
     * @dev Sets a new implementation address
     * @param newImplementationAddress_: address of the contract with the new implementation
     */
    function setImplementation(address newImplementationAddress_) public onlyOwner {
        _implementation = newImplementationAddress_;
    }

    /**
     * @dev Add a new address and "pseudo" salt to the externalContractRegistry
     * @param salt_: salt to be used to retrieve the contract
     * @param newContractAddress_: address of the contract to be added to registry
     */
    function addNewExternalContract(bytes32 salt_, address newContractAddress_) public onlyOwner {
        if (_externalContractRegistry[salt_].exist) {
            revert AliceNetFactoryBaseErrors.SaltAlreadyInUse();
        }
        _externalContractRegistry[salt_] = ContractInfo(true, newContractAddress_);
    }

    /**
     * @dev Sets the new owner
     * @param newOwner_: address of the new owner
     */
    function setOwner(address newOwner_) public onlyOwner {
        _owner = newOwner_;
    }

    /**
     * @dev lookup allows anyone interacting with the contract to get the address of contract specified
     * by its salt_
     * @param salt_: Custom NatSpec tag @custom:salt at the top of the contract solidity file
     */
    function lookup(bytes32 salt_) public view returns (address) {
        // check if the salt belongs to one of the pre-defined contracts deployed during the factory deployment
        if (salt_ == _aTokenSalt) {
            return _aTokenAddress;
        }
        // check if the salt belongs to any address in the external contract registry (contracts deployed outside the factory)
        ContractInfo memory contractInfo = _externalContractRegistry[salt_];
        if (contractInfo.exist) {
            return contractInfo.logicAddr;
        }
        return getMetamorphicContractAddress(salt_, address(this));
    }

    /**
     * @dev getImplementation is public getter function for the _owner account address
     */
    function getImplementation() public view returns (address) {
        return _implementation;
    }

    /**
     * @dev owner is public getter function for the _owner account address
     * @return owner_ address of the owner account
     */
    function owner() public view returns (address owner_) {
        owner_ = _owner;
    }

    /**
     * @dev contracts is public getter that gets the array of salts associated with all the contracts
     * deployed with this factory
     * @return contracts_ the array of salts associated with all the contracts deployed with this
     * factory
     */
    function contracts() public view returns (bytes32[] memory contracts_) {
        contracts_ = _contracts;
    }

    /**
     * @dev getNumContracts getter function for retrieving the total number of contracts
     * deployed with this factory
     * @return the length of the contract array
     */
    function getNumContracts() public view returns (uint256) {
        return _contracts.length;
    }

    /**
     * @dev _callAny allows EOA to call function impersonating the factory address
     * @param target_: the address of the contract to be called
     * @param value_: value in WEIs to send together the call
     * @param cdata_: Hex encoded data with function signature + arguments of the target function to be called
     */
    function _callAny(
        address target_,
        uint256 value_,
        bytes memory cdata_
    ) internal {
        assembly {
            let size := mload(cdata_)
            let ptr := add(0x20, cdata_)
            if iszero(call(gas(), target_, value_, ptr, size, 0x00, 0x00)) {
                returndatacopy(0x00, 0x00, returndatasize())
                revert(0x00, returndatasize())
            }
        }
    }

    /**
     * @dev _delegateCallAny allows EOA to call a function in a contract without impersonating the factory
     * @param target_: the address of the contract to be called
     * @param cdata_: Hex encoded data with function signature + arguments of the target function to be called
     */
    function _delegateCallAny(address target_, bytes memory cdata_) internal {
        assembly {
            let size := mload(cdata_)
            let ptr := add(0x20, cdata_)
            if iszero(delegatecall(gas(), target_, ptr, size, 0x00, 0x00)) {
                returndatacopy(0x00, 0x00, returndatasize())
                revert(0x00, returndatasize())
            }
        }
    }

    /**
     * @dev _deployCreate allows the owner to deploy raw contracts through the factory using
     * non-deterministic address generation (create OpCode)
     * @param deployCode_ Hex encoded data with the deployment code of the contract to be deployed +
     * constructors' args (if any)
     * @return contractAddr the deployed contract address
     */
    function _deployCreate(bytes calldata deployCode_) internal returns (address contractAddr) {
        assembly {
            //get the next free pointer
            let basePtr := mload(0x40)
            let ptr := basePtr

            //copies the initialization code of the implementation contract
            calldatacopy(ptr, deployCode_.offset, deployCode_.length)

            //Move the ptr to the end of the code in memory
            ptr := add(ptr, deployCode_.length)

            contractAddr := create(0, basePtr, sub(ptr, basePtr))
        }
        _codeSizeZeroRevert((_extCodeSize(contractAddr) != 0));
        emit DeployedRaw(contractAddr);
        return contractAddr;
    }

    /**
     * @dev _deployCreate2 allows the owner to deploy contracts with deterministic address through the
     * factory
     * @param value_ endowment value in WEIS for the created contract
     * @param salt_ salt used to determine the final determinist address for the deployed contract
     * @param deployCode_ Hex encoded data with the deployment code of the contract to be deployed +
     * constructors' args (if any)
     * @return contractAddr the deployed contract address
     */
    function _deployCreate2(
        uint256 value_,
        bytes32 salt_,
        bytes calldata deployCode_
    ) internal returns (address contractAddr) {
        assembly {
            //get the next free pointer
            let basePtr := mload(0x40)
            let ptr := basePtr

            //copies the initialization code of the implementation contract
            calldatacopy(ptr, deployCode_.offset, deployCode_.length)

            //Move the ptr to the end of the code in memory
            ptr := add(ptr, deployCode_.length)

            contractAddr := create2(value_, basePtr, sub(ptr, basePtr), salt_)
        }
        _codeSizeZeroRevert(uint160(contractAddr) != 0);
        //record the contract salt to the _contracts array for lookup
        _contracts.push(salt_);
        emit DeployedRaw(contractAddr);
        return contractAddr;
    }

    /**
     * @dev _deployProxy deploys a proxy contract with upgradable logic. See Proxy.sol contract.
     * @param salt_ salt used to determine the final determinist address for the deployed contract
     */
    function _deployProxy(bytes32 salt_) internal returns (address contractAddr) {
        address proxyTemplate = _proxyTemplate;
        assembly {
            // store proxy template address as implementation,
            sstore(_implementation.slot, proxyTemplate)
            let ptr := mload(0x40)
            mstore(0x40, add(ptr, 0x20))
            // put metamorphic code as initCode
            // push1 20
            mstore(ptr, shl(72, 0x6020363636335afa1536363636515af43d36363e3d36f3))
            contractAddr := create2(0, ptr, 0x17, salt_)
        }
        _codeSizeZeroRevert((_extCodeSize(contractAddr) != 0));
        // record the contract salt to the contracts array
        _contracts.push(salt_);
        emit DeployedProxy(contractAddr);
        return contractAddr;
    }

    /**
     * @dev _initializeContract allows the owner/delegator to initialize contracts deployed via factory
     * @param contract_ address of the contract that will be initialized
     * @param initCallData_ Hex encoded initialization function signature + parameters to initialize the
     * deployed contract
     */
    function _initializeContract(address contract_, bytes calldata initCallData_) internal {
        assembly {
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
     * @dev _multiCall allows EOA to make multiple function calls within a single transaction
     * impersonating the factory
     * @param cdata_: array of abi encoded data with the function calls (function signature + arguments)
     */
    function _multiCall(MultiCallArgs[] calldata cdata_) internal {
        for (uint256 i = 0; i < cdata_.length; i++) {
            _callAny(cdata_[i].target, cdata_[i].value, cdata_[i].data);
        }
        _returnAvailableData();
    }

    /**
     * @dev _upgradeProxy updates the implementation/logic address of an already deployed proxy contract.
     * @param salt_ salt used to determine the final determinist address for the deployed proxy contract
     * @param newImpl_ address of the new contract that contains the new implementation logic
     * @param initCallData_ Hex encoded initialization function signature + parameters to initialize the
     * new implementation contract
     */
    function _upgradeProxy(
        bytes32 salt_,
        address newImpl_,
        bytes calldata initCallData_
    ) internal {
        address proxy = DeterministicAddress.getMetamorphicContractAddress(salt_, address(this));
        __upgrade(proxy, newImpl_);
        assert(IProxy(proxy).getImplementationAddress() == newImpl_);
        _initializeContract(proxy, initCallData_);
    }

    /**
     * @dev Aux function to return the external code size
     */
    function _extCodeSize(address target_) internal view returns (uint256 size) {
        assembly {
            size := extcodesize(target_)
        }
        return size;
    }

    /**
     * @dev Aux function to get the return data size
     */
    function _returnAvailableData() internal pure {
        assembly {
            returndatacopy(0x00, 0x00, returndatasize())
            return(0x00, returndatasize())
        }
    }

    /**
     * @dev _requireAuth reverts if false and returns unauthorized error message
     * @param isOk_ boolean false to cause revert
     */
    function _requireAuth(bool isOk_) internal pure {
        if (!isOk_) {
            revert AliceNetFactoryBaseErrors.Unauthorized();
        }
    }

    /**
     * @dev _codeSizeZeroRevert reverts if false and returns csize0 error message
     * @param isOk_ boolean false to cause revert
     */
    function _codeSizeZeroRevert(bool isOk_) internal pure {
        if (!isOk_) {
            revert AliceNetFactoryBaseErrors.CodeSizeZero();
        }
    }
}
