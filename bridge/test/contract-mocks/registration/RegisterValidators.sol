// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/ISnapshots.sol";
import "contracts/interfaces/IStakingToken.sol";
import "contracts/interfaces/IUtilityToken.sol";
import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/interfaces/IStakingNFT.sol";
import "contracts/interfaces/IETHDKG.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";
import "@openzeppelin/contracts/token/ERC721/IERC721.sol";

contract ExternalStoreRegistration is ImmutableFactory {
    uint256 internal _counter;
    uint256[] internal _tokenIDs;

    constructor(address factory_) ImmutableFactory(factory_) {}

    function storeTokenIds(uint256[] memory tokenIDs) public onlyFactory {
        _tokenIDs = tokenIDs;
    }

    function incrementCounter() public onlyFactory {
        _counter++;
    }

    function getTokenIds() public view returns (uint256[] memory) {
        uint256[] memory ret = new uint256[](_tokenIDs.length);
        for (uint256 i = 0; i < _tokenIDs.length; i++) {
            ret[i] = _tokenIDs[i];
        }
        return ret;
    }

    function getCounter() public view returns (uint256) {
        return _counter;
    }

    function getTokenIDsLength() public view returns (uint256) {
        return _tokenIDs.length;
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
        IStakingTokenMinter(_aTokenMinterAddress()).mint(_factoryAddress(), numValidators);
        IERC20Transferable(_aTokenAddress()).approve(_publicStakingAddress(), numValidators);
        uint256[] memory tokenIDs = new uint256[](numValidators);
        for (uint256 i; i < numValidators; i++) {
            // minting publicStaking position for the factory
            tokenIDs[i] = IStakingNFT(_publicStakingAddress()).mint(1);
            IERC721(_publicStakingAddress()).approve(_validatorPoolAddress(), tokenIDs[i]);
        }
        _externalStore.storeTokenIds(tokenIDs);
    }

    function registerValidators(address[] calldata validatorsAccounts_) public {
        require(
            validatorsAccounts_.length == _externalStore.getTokenIDsLength(),
            "Incorrect validators account length!"
        );
        uint256[] memory tokenIDs = _externalStore.getTokenIds();
        ////////////// Registering validators /////////////////////////
        IValidatorPool(_validatorPoolAddress()).registerValidators(validatorsAccounts_, tokenIDs);
    }
}
