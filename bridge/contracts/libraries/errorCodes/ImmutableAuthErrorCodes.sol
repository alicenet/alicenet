// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library ImmutableAuthErrorCodes {
    // ImmutableAuth error codes
    bytes32 public constant IMMUTEABLEAUTH_ONLY_FACTORY = "2000"; //"onlyFactory"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_ATOKEN = "2001"; //"onlyAToken"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_FOUNDATION = "2002"; //"onlyFoundation"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_GOVERNANCE = "2003"; // "onlyGovernance"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_LIQUIDITYPROVIDERSTAKING = "2004"; // "onlyLiquidityProviderStaking"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_BTOKEN = "2005"; // "onlyBToken"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_MADTOKEN = "2006"; // "onlyMadToken"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_PUBLICSTAKING = "2007"; // "onlyPublicStaking"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_SNAPSHOTS = "2008"; // "onlySnapshots"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_STAKINGPOSITIONDESCRIPTOR = "2009"; // "onlyStakingPositionDescriptor"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_VALIDATORPOOL = "2010"; // "onlyValidatorPool"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_VALIDATORSTAKING = "2011"; // "onlyValidatorStaking"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_ATOKENBURNER = "2012"; // "onlyATokenBurner"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_ATOKENMINTER = "2013"; // "onlyATokenMinter"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_ETHDKGACCUSATIONS = "2014"; // "onlyETHDKGAccusations"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_ETHDKGPHASES = "2015"; // "onlyETHDKGPhases"
    bytes32 public constant IMMUTEABLEAUTH_ONLY_ETHDKG = "2016"; // "onlyETHDKG"
}
