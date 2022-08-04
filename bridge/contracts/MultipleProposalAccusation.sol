// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/ISnapshots.sol";
import "contracts/libraries/parsers/PClaimsParserLibrary.sol";
import "contracts/libraries/math/CryptoLibrary.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/utils/AccusationsLibrary.sol";

/// @custom:salt MultipleProposalAccusation
/// @custom:deploy-type deployUpgradeable
/// @custom:salt-type Accusation
contract MultipleProposalAccusation is
    ImmutableFactory,
    ImmutableSnapshots,
    ImmutableETHDKG,
    ImmutableValidatorPool
{
    mapping(bytes32 => bool) internal _accusations;

    constructor()
        ImmutableFactory(msg.sender)
        ImmutableSnapshots()
        ImmutableETHDKG()
        ImmutableValidatorPool()
    {}

    function calculateId(bytes calldata signature0_,
        bytes calldata pClaims0_,
        bytes calldata signature1_,
        bytes calldata pClaims1_) public pure returns(bytes32) {
        // convert signatures to 32byte uints
        uint256 sig0 = uint256(keccak256(signature0_));
        uint256 sig1 = uint256(keccak256(signature1_));
        bytes32 id;

        // check sorting of signatures to generate ID
        if (sig0 <= sig1) {
            id = keccak256(abi.encode(signature0_, pClaims0_, signature1_, pClaims1_));
        } else {
            id = keccak256(abi.encode(signature1_, pClaims1_, signature0_, pClaims0_));
        }

        return id;
    }

    /// @notice This function validates an accusation of multiple proposals.
    /// @param signature0_ The signature of pclaims0
    /// @param pClaims0_ The PClaims of the accusation
    /// @param signature1_ The signature of pclaims1
    /// @param pClaims1_ The PClaims of the accusation
    /// @return the address of the signer
    function accuseMultipleProposal(
        bytes calldata signature0_,
        bytes calldata pClaims0_,
        bytes calldata signature1_,
        bytes calldata pClaims1_
    ) public returns (address) {
        // convert signatures to 32byte uints
        uint256 sig0 = uint256(keccak256(signature0_));
        uint256 sig1 = uint256(keccak256(signature1_));
        bytes32 id;

        // check sorting of signatures to generate ID
        if (sig0 <= sig1) {
            id = keccak256(abi.encode(signature0_, pClaims0_, signature1_, pClaims1_));
        } else {
            id = keccak256(abi.encode(signature1_, pClaims1_, signature0_, pClaims0_));
        }

        // check if thi accusation has already been submitted
        require(!_accusations[id], "Accusations: the accusation has already been submitted!");

        // ecrecover sig0/1 and ensure both are valid and accounts are equal
        address signerAccount0 = AccusationsLibrary.recoverMadNetSigner(signature0_, pClaims0_);
        address signerAccount1 = AccusationsLibrary.recoverMadNetSigner(signature1_, pClaims1_);

        require(
            signerAccount0 == signerAccount1,
            "Accusations: the signers of the proposals should be the same"
        );

        // ensure the hashes of blob0/1 are different
        require(
            keccak256(pClaims0_) != keccak256(pClaims1_),
            "Accusations: the PClaims are equal!"
        );

        PClaimsParserLibrary.PClaims memory pClaims0 = PClaimsParserLibrary.extractPClaims(
            pClaims0_
        );
        PClaimsParserLibrary.PClaims memory pClaims1 = PClaimsParserLibrary.extractPClaims(
            pClaims1_
        );

        // ensure the height of blob0/1 are equal using RCert sub object of PClaims
        require(
            pClaims0.rCert.rClaims.height == pClaims1.rCert.rClaims.height,
            "Accusations: the block heights between the proposals are different!"
        );

        // ensure the round of blob0/1 are equal using RCert sub object of PClaims
        require(
            pClaims0.rCert.rClaims.round == pClaims1.rCert.rClaims.round,
            "Accusations: the round between the proposals are different!"
        );

        // ensure the chainid of blob0/1 are equal using RCert sub object of PClaims
        require(
            pClaims0.rCert.rClaims.chainId == pClaims1.rCert.rClaims.chainId,
            "Accusations: the chainId between the proposals are different!"
        );

        // ensure the chainid of blob0 is correct for this chain using RCert sub object of PClaims
        uint256 chainId = ISnapshots(_snapshotsAddress()).getChainId();
        require(
            pClaims0.rCert.rClaims.chainId == chainId,
            "Accusations: the chainId is invalid for this chain!"
        );

        _accusations[id] = true;

        // major slash this validator. Note: this method already checks if the dishonestValidator (1st argument) is a validator.
        IValidatorPool(_validatorPoolAddress()).majorSlash(signerAccount0, msg.sender);

        return signerAccount0;
    }

    function isAccused(bytes32 id_) public view returns (bool) {
        return _accusations[id_];
    }
}
