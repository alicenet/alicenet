// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/ISnapshots.sol";
import "contracts/interfaces/IAToken.sol";
import "contracts/interfaces/IBToken.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/interfaces/IStakingNFT.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";

import "hardhat/console.sol";

contract ExternalStoreRegistration is ImmutableFactory {
    uint256[4] internal _tokenIDs;
    uint256 internal _counter;

    constructor(address factory_) ImmutableFactory(factory_) {}

    function storeTokenIds(uint256[4] memory tokenIDs) public onlyFactory {
        _tokenIDs = tokenIDs;
    }

    function incrementCounter() public onlyFactory {
        _counter++;
    }

    function getTokenIds() public view returns (uint256[4] memory) {
        return _tokenIDs;
    }

    function getCounter() public view returns (uint256) {
        return _counter;
    }
}

contract RegisterValidators is
    ImmutableFactory,
    ImmutableSnapshots,
    ImmutableETHDKG,
    ImmutableAToken,
    ImmutableATokenMinter,
    ImmutableBToken,
    ImmutablePublicStaking,
    ImmutableValidatorPool
{
    uint256 public constant EPOCH_LENGTH = 1024;
    ExternalStoreRegistration internal immutable _externalStore;

    constructor(address factory_)
        ImmutableFactory(factory_)
        ImmutableSnapshots()
        ImmutableETHDKG()
        ImmutableAToken()
        ImmutableBToken()
        ImmutableATokenMinter()
        ImmutablePublicStaking()
        ImmutableValidatorPool()
    {
        _externalStore = new ExternalStoreRegistration(factory_);
    }

    function stakeValidators(uint256 numValidators) public {
        // Setting staking amount
        IValidatorPool(_validatorPoolAddress()).setStakeAmount(1);
        // Minting 4 aTokensWei to stake the validators
        IATokenMinter(_aTokenMinterAddress()).mint(_factoryAddress(), numValidators);
        IERC20Transferable(_aTokenAddress()).approve(_publicStakingAddress(), numValidators);
        uint256[4] memory tokenIDs;
        for (uint256 i; i < numValidators; i++) {
            // minting publicStaking position for the factory
            tokenIDs[i] = IStakingNFT(_publicStakingAddress()).mint(1);
            IERC721(_publicStakingAddress()).approve(_validatorPoolAddress(), tokenIDs[i]);
        }
        _externalStore.storeTokenIds(tokenIDs);
    }

    function registerValidators(address[] calldata validatorsAccounts_) public {
        uint256[] memory tokenIDs = new uint256[](4);
        uint256[4] memory tokenIDs_ = _externalStore.getTokenIds();
        for (uint256 i = 0; i < tokenIDs.length; i++) {
            tokenIDs[i] = tokenIDs_[i];
        }
        ////////////// Registering validators /////////////////////////
        IValidatorPool(_validatorPoolAddress()).registerValidators(validatorsAccounts_, tokenIDs);
    }
}
