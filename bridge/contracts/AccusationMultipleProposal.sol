// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;

import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/ISnapshots.sol";
import "contracts/libraries/parsers/PClaimsParserLibrary.sol";
import "contracts/libraries/math/CryptoLibrary.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/utils/AccusationsLibrary.sol";
import "contracts/libraries/errors/AccusationsErrors.sol";

/// @custom:deploy-type deployUpgradeable
/// @custom:role Accusation
/// @custom:salt AccusationMultipleProposal
contract AccusationMultipleProposal is
    ImmutableFactory,
    ImmutableSnapshots,
    ImmutableETHDKG,
    ImmutableValidatorPool
{
    // this is the keccak256 of "AccusationMultipleProposal"
    bytes32 public constant PRE_SALT =
        0x17287210c71008320429d4cce2075373f0b2c5217b507513fe4904fead741aad;
    mapping(bytes32 => bool) internal _accusations;

    constructor()
        ImmutableFactory(msg.sender)
        ImmutableSnapshots()
        ImmutableETHDKG()
        ImmutableValidatorPool()
    {}

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
        // ecrecover sig0/1 and ensure both are valid and accounts are equal
        address signerAccount0 = AccusationsLibrary.recoverMadNetSigner(signature0_, pClaims0_);
        address signerAccount1 = AccusationsLibrary.recoverMadNetSigner(signature1_, pClaims1_);

        if (signerAccount0 != signerAccount1) {
            revert AccusationsErrors.SignersDoNotMatch(signerAccount0, signerAccount1);
        }

        // ensure the hashes of blob0/1 are different
        if (keccak256(pClaims0_) == keccak256(pClaims1_)) {
            revert AccusationsErrors.PClaimsAreEqual();
        }

        PClaimsParserLibrary.PClaims memory pClaims0 = PClaimsParserLibrary.extractPClaims(
            pClaims0_
        );
        PClaimsParserLibrary.PClaims memory pClaims1 = PClaimsParserLibrary.extractPClaims(
            pClaims1_
        );

        // ensure the height of blob0/1 are equal using RCert sub object of PClaims
        if (pClaims0.rCert.rClaims.height != pClaims1.rCert.rClaims.height) {
            revert AccusationsErrors.PClaimsHeightsDoNotMatch(
                pClaims0.rCert.rClaims.height,
                pClaims1.rCert.rClaims.height
            );
        }

        // ensure the round of blob0/1 are equal using RCert sub object of PClaims
        if (pClaims0.rCert.rClaims.round != pClaims1.rCert.rClaims.round) {
            revert AccusationsErrors.PClaimsRoundsDoNotMatch(
                pClaims0.rCert.rClaims.round,
                pClaims1.rCert.rClaims.round
            );
        }

        // ensure the chainid of blob0/1 are equal using RCert sub object of PClaims
        if (pClaims0.rCert.rClaims.chainId != pClaims1.rCert.rClaims.chainId) {
            revert AccusationsErrors.PClaimsChainIdsDoNotMatch(
                pClaims0.rCert.rClaims.chainId,
                pClaims1.rCert.rClaims.chainId
            );
        }

        // ensure the chainid of blob0 is correct for this chain using RCert sub object of PClaims
        uint256 chainId = ISnapshots(_snapshotsAddress()).getChainId();
        if (pClaims0.rCert.rClaims.chainId != chainId) {
            revert AccusationsErrors.InvalidChainId(pClaims0.rCert.rClaims.chainId, chainId);
        }

        // deterministic accusation ID
        bytes32 id = keccak256(
            abi.encodePacked(
                signerAccount0,
                pClaims0.rCert.rClaims.chainId,
                pClaims0.rCert.rClaims.height,
                pClaims0.rCert.rClaims.round,
                PRE_SALT
            )
        );

        // check if this accusation ID has already been submitted
        if (_accusations[id]) {
            revert AccusationsErrors.AccusationAlreadySubmitted(id);
        }

        _accusations[id] = true;

        // major slash this validator. Note: this method already checks if the dishonest validator (1st argument) is an accusable validator.
        IValidatorPool(_validatorPoolAddress()).majorSlash(signerAccount0, msg.sender, PRE_SALT);

        return signerAccount0;
    }

    /// @notice This function tells whether an accusation ID has already been submitted or not.
    /// @param id_ The deterministic accusation ID
    /// @return true if the ID has already been submitted, false otherwise
    function isAccused(bytes32 id_) public view returns (bool) {
        return _accusations[id_];
    }
}
