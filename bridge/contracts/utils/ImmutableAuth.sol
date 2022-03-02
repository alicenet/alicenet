
// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "./DeterministicAddress.sol";

abstract contract ImmutableFactory is DeterministicAddress {

    address private immutable _factory;

    constructor(address factory_) {
        _factory = factory_;
    }

    modifier onlyFactory() {
        require(msg.sender == _factory, "onlyFactory");
        _;
    }

    function _factoryAddress() internal view returns(address) {
        return _factory;
    }

}


abstract contract ImmutableValidatorNFT is ImmutableFactory {

    address private immutable _ValidatorNFT;

    constructor() {
        _ValidatorNFT = getMetamorphicContractAddress(0x56616c696461746f724e46540000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyValidatorNFT() {
        require(msg.sender == _ValidatorNFT, "onlyValidatorNFT");
        _;
    }

    function _ValidatorNFTAddress() internal view returns(address) {
        return _ValidatorNFT;
    }

    function _saltForValidatorNFT() internal pure returns(bytes32) {
        return 0x56616c696461746f724e46540000000000000000000000000000000000000000;
    }

}



abstract contract ImmutableMadToken is ImmutableFactory {

    address private immutable _MadToken;

    constructor() {
        _MadToken = getMetamorphicContractAddress(0x4d6164546f6b656e000000000000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyMadToken() {
        require(msg.sender == _MadToken, "onlyMadToken");
        _;
    }

    function _MadTokenAddress() internal view returns(address) {
        return _MadToken;
    }

    function _saltForMadToken() internal pure returns(bytes32) {
        return 0x4d6164546f6b656e000000000000000000000000000000000000000000000000;
    }

}



abstract contract ImmutableStakeNFT is ImmutableFactory {

    address private immutable _StakeNFT;

    constructor() {
        _StakeNFT = getMetamorphicContractAddress(0x5374616b654e4654000000000000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyStakeNFT() {
        require(msg.sender == _StakeNFT, "onlyStakeNFT");
        _;
    }

    function _StakeNFTAddress() internal view returns(address) {
        return _StakeNFT;
    }

    function _saltForStakeNFT() internal pure returns(bytes32) {
        return 0x5374616b654e4654000000000000000000000000000000000000000000000000;
    }

}



abstract contract ImmutableMadByte is ImmutableFactory {

    address private immutable _MadByte;

    constructor() {
        _MadByte = getMetamorphicContractAddress(0x4d61644279746500000000000000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyMadByte() {
        require(msg.sender == _MadByte, "onlyMadByte");
        _;
    }

    function _MadByteAddress() internal view returns(address) {
        return _MadByte;
    }

    function _saltForMadByte() internal pure returns(bytes32) {
        return 0x4d61644279746500000000000000000000000000000000000000000000000000;
    }

}



abstract contract ImmutableGovernance is ImmutableFactory {

    address private immutable _Governance;

    constructor() {
        _Governance = getMetamorphicContractAddress(0x476f7665726e616e636500000000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyGovernance() {
        require(msg.sender == _Governance, "onlyGovernance");
        _;
    }

    function _GovernanceAddress() internal view returns(address) {
        return _Governance;
    }

    function _saltForGovernance() internal pure returns(bytes32) {
        return 0x476f7665726e616e636500000000000000000000000000000000000000000000;
    }

}



abstract contract ImmutableValidatorPool is ImmutableFactory {

    address private immutable _ValidatorPool;

    constructor() {
        _ValidatorPool = getMetamorphicContractAddress(0x56616c696461746f72506f6f6c00000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyValidatorPool() {
        require(msg.sender == _ValidatorPool, "onlyValidatorPool");
        _;
    }

    function _ValidatorPoolAddress() internal view returns(address) {
        return _ValidatorPool;
    }

    function _saltForValidatorPool() internal pure returns(bytes32) {
        return 0x56616c696461746f72506f6f6c00000000000000000000000000000000000000;
    }

}



abstract contract ImmutableETHDKG is ImmutableFactory {

    address private immutable _ETHDKG;

    constructor() {
        _ETHDKG = getMetamorphicContractAddress(0x455448444b470000000000000000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyETHDKG() {
        require(msg.sender == _ETHDKG, "onlyETHDKG");
        _;
    }

    function _ETHDKGAddress() internal view returns(address) {
        return _ETHDKG;
    }

    function _saltForETHDKG() internal pure returns(bytes32) {
        return 0x455448444b470000000000000000000000000000000000000000000000000000;
    }

}



abstract contract ImmutableETHDKGAccusations is ImmutableFactory {

    address private immutable _ETHDKGAccusations;

    constructor() {
        _ETHDKGAccusations = getMetamorphicContractAddress(0x455448444b4741636375736174696f6e73000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyETHDKGAccusations() {
        require(msg.sender == _ETHDKGAccusations, "onlyETHDKGAccusations");
        _;
    }

    function _ETHDKGAccusationsAddress() internal view returns(address) {
        return _ETHDKGAccusations;
    }

    function _saltForETHDKGAccusations() internal pure returns(bytes32) {
        return 0x455448444b4741636375736174696f6e73000000000000000000000000000000;
    }

}



abstract contract ImmutableSnapshots is ImmutableFactory {

    address private immutable _Snapshots;

    constructor() {
        _Snapshots = getMetamorphicContractAddress(0x536e617073686f74730000000000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlySnapshots() {
        require(msg.sender == _Snapshots, "onlySnapshots");
        _;
    }

    function _SnapshotsAddress() internal view returns(address) {
        return _Snapshots;
    }

    function _saltForSnapshots() internal pure returns(bytes32) {
        return 0x536e617073686f74730000000000000000000000000000000000000000000000;
    }

}



abstract contract ImmutableETHDKGPhases is ImmutableFactory {

    address private immutable _ETHDKGPhases;

    constructor() {
        _ETHDKGPhases = getMetamorphicContractAddress(0x455448444b475068617365730000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyETHDKGPhases() {
        require(msg.sender == _ETHDKGPhases, "onlyETHDKGPhases");
        _;
    }

    function _ETHDKGPhasesAddress() internal view returns(address) {
        return _ETHDKGPhases;
    }

    function _saltForETHDKGPhases() internal pure returns(bytes32) {
        return 0x455448444b475068617365730000000000000000000000000000000000000000;
    }

}



abstract contract ImmutableStakeNFTLP is ImmutableFactory {

    address private immutable _StakeNFTLP;

    constructor() {
        _StakeNFTLP = getMetamorphicContractAddress(0x5374616b654e46544c5000000000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyStakeNFTLP() {
        require(msg.sender == _StakeNFTLP, "onlyStakeNFTLP");
        _;
    }

    function _StakeNFTLPAddress() internal view returns(address) {
        return _StakeNFTLP;
    }

    function _saltForStakeNFTLP() internal pure returns(bytes32) {
        return 0x5374616b654e46544c5000000000000000000000000000000000000000000000;
    }

}



abstract contract ImmutableFoundation is ImmutableFactory {

    address private immutable _Foundation;

    constructor() {
        _Foundation = getMetamorphicContractAddress(0x466f756e646174696f6e00000000000000000000000000000000000000000000, _factoryAddress());
    }

    modifier onlyFoundation() {
        require(msg.sender == _Foundation, "onlyFoundation");
        _;
    }

    function _FoundationAddress() internal view returns(address) {
        return _Foundation;
    }

    function _saltForFoundation() internal pure returns(bytes32) {
        return 0x466f756e646174696f6e00000000000000000000000000000000000000000000;
    }

}


