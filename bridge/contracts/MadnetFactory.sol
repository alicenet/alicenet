// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;
import "contracts/utils/DeterministicAddress.sol";
import "contracts/Proxy.sol";
import "contracts/libraries/factory/MadnetFactoryBase.sol";


/// @custom:salt MadnetFactory
/// @custom:deploy-type factory
contract MadnetFactory is MadnetFactoryBase {
    /**
     * @dev The constructor encodes the proxy deploy byte code with the universal deploycode at
     * the head and the factory address at the tail, and deploys the proxy byte code using create
     * The result of this deployment will be a contract with the proxy contract deployment bytecode
     * with its constructor at the head, runtime code in the body and constructor args at the tail.
     * the constructor then sets proxyTemplate_ state var to the deployed proxy template address
     * the deploy account will be set as the first owner of the factory.
     * @param selfAddr_ is the factory contracts address (address of itself)
     */
    constructor(address selfAddr_) MadnetFactoryBase(selfAddr_) {}

    /**
     * @dev deployCreate allows the owner to deploy raw contracts through the factory
     * @param _deployCode bytecode to deploy using create
     */
    function deployCreate(bytes calldata _deployCode)
        public
        onlyOwner
        returns (address contractAddr)
    {
        return deployCreateInternal(_deployCode);
    }

    /**
     * @dev deployCreate2 allows the owner to deploy contracts with deterministic address
     * through the factory
     * @param _value endowment value for created contract
     * @param _salt salt for create2 deployment, used to distinguish contracts deployed from this factory
     * @param _deployCode bytecode to deploy using create2
     */
    function deployCreate2(
        uint256 _value,
        bytes32 _salt,
        bytes calldata _deployCode
    ) public payable onlyOwner returns (address contractAddr) {
        contractAddr = deployCreate2Internal(_value, _salt, _deployCode);
    }

    function initializeContract(address _contract, bytes calldata _initCallData)
        public
        onlyOwnerOrDelegator
    {
        initializeContractInternal(_contract, _initCallData);
    }

    /**
     * @dev deployTemplate deploys a template contract with the universal code copy constructor that deploys
     * the deploycode as the contracts runtime code.
     * @param _deployCode the deploycode with the constructor args appended if any
     * @return contractAddr the address of the deployed template contract
     */
    function deployTemplate(bytes calldata _deployCode)
        public
        onlyOwner
        returns (address contractAddr)
    {
        contractAddr = deployTemplateInternal(_deployCode);
    }

    /**
     * @dev deployStatic deploys a template contract with the universal code copy constructor that deploys
     * the deploycode as the contracts runtime code.
     * @param _salt sda
     * @return contractAddr the address of the deployed template contract
     */
    function deployStatic(bytes32 _salt, bytes calldata _initCallData)
        public
        onlyOwner
        returns (address contractAddr)
    {
        contractAddr = deployStaticInternal(_salt, _initCallData);
    }

    /**
     *
     */
    function deployProxy(bytes32 _salt) public onlyOwner returns (address contractAddr) {
        contractAddr = deployProxyInternal(_salt);
    }

    function upgradeProxy(
        bytes32 _salt,
        address _newImpl,
        bytes calldata _initCallData
    ) public onlyOwner {
        upgradeProxyInternal(_salt, _newImpl, _initCallData);
    }

    /**
     * @dev delegateCallAny is a access restricted wrapper function for delegateCallAnyInternal
     * @param _target: the address of the contract to call
     * @param _cdata: the call data for the delegate call
     */
    function delegateCallAny(address _target, bytes calldata _cdata) public onlyOwnerOrDelegator {
        bytes memory cdata = _cdata;
        delegateCallAnyInternal(_target, cdata);
        returnAvailableData();
    }

    /**
     * @dev callAny is a access restricted wrapper function for callAnyInternal
     * @param _target: the address of the contract to call
     * @param _value: value to send with the call
     * @param _cdata: the call data for the delegate call
     */
    function callAny(
        address _target,
        uint256 _value,
        bytes calldata _cdata
    ) public onlyOwner {
        bytes memory cdata = _cdata;
        callAnyInternal(_target, _value, cdata);
        returnAvailableData();
    }

    /**
     * @dev multiCall allows EOA to make multiple function calls within a single transaction **in this contract**, and only returns the result of the last call
     * @param _cdata: array of function calls
     * returns the result of the last call
     */
    function multiCall(bytes[] calldata _cdata) public onlyOwner {
        multiCallInternal(_cdata);
    }
}
