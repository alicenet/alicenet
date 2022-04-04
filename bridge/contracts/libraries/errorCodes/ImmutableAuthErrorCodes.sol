// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

library ImmutableAuthErrorCodes {
    // ImmutableAuth error codes
    uint16 public constant IMMUTEABLEAUTH_ONLY_FACTORY = 2000; //"onlyFactory"
    uint16 public constant IMMUTEABLEAUTH_ONLY_ATOKEN = 2001; //"onlyAToken"
    uint16 public constant IMMUTEABLEAUTH_ONLY_FOUNDATION = 2002; //"onlyFoundation"
    uint16 public constant IMMUTEABLEAUTH_ONLY_GOVERNANCE = 2003; // "onlyGovernance"
    uint16 public constant IMMUTEABLEAUTH_ONLY_LIQUIDITYPROVIDERSTAKING = 2004; // "onlyLiquidityProviderStaking"
    uint16 public constant IMMUTEABLEAUTH_ONLY_BTOKEN = 2005; // "onlyBToken"
    uint16 public constant IMMUTEABLEAUTH_ONLY_MADTOKEN = 2006; // "onlyMadToken"
    uint16 public constant IMMUTEABLEAUTH_ONLY_PUBLICSTAKING = 2007; // "onlyPublicStaking"
    uint16 public constant IMMUTEABLEAUTH_ONLY_SNAPSHOTS = 2008; // "onlySnapshots"
    uint16 public constant IMMUTEABLEAUTH_ONLY_STAKINGPOSITIONDESCRIPTOR = 2009; // "onlyStakingPositionDescriptor"
    uint16 public constant IMMUTEABLEAUTH_ONLY_VALIDATORPOOL = 2010; // "onlyValidatorPool"
    uint16 public constant IMMUTEABLEAUTH_ONLY_VALIDATORSTAKING = 2011; // "onlyValidatorStaking"
    uint16 public constant IMMUTEABLEAUTH_ONLY_ATOKENBURNER = 2012; // "onlyATokenBurner"
    uint16 public constant IMMUTEABLEAUTH_ONLY_ATOKENMINTER = 2013; // "onlyATokenMinter"
    uint16 public constant IMMUTEABLEAUTH_ONLY_ETHDKGACCUSATIONS = 2014; // "onlyETHDKGAccusations"
    uint16 public constant IMMUTEABLEAUTH_ONLY_ETHDKGPHASES = 2015; // "onlyETHDKGPhases"
    uint16 public constant IMMUTEABLEAUTH_ONLY_ETHDKG = 2016; // "onlyETHDKG"
}
