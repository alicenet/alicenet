// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.11;

import "contracts/interfaces/IValidatorPool.sol";
import "contracts/interfaces/ISnapshots.sol";
import "contracts/libraries/parsers/PClaimsParserLibrary.sol";
import "contracts/libraries/math/CryptoLibrary.sol";
import "contracts/utils/ImmutableAuth.sol";
import "contracts/utils/AccusationsLibrary.sol";

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

    /// @notice This function validates an accusation of multiple proposals.
    /// @param _signature0 The signature of pclaims0
    /// @param _pClaims0 The PClaims of the accusation
    /// @param _signature1 The signature of pclaims1
    /// @param _pClaims1 The PClaims of the accusation
    /// @return the address of the signer
    function AccuseMultipleProposal(
        bytes calldata _signature0,
        bytes calldata _pClaims0,
        bytes calldata _signature1,
        bytes calldata _pClaims1
    ) internal view returns (address) {
        // ecrecover sig0/1 and ensure both are valid and accounts are equal
        address signerAccount0 = AccusationsLibrary.recoverMadNetSigner(_signature0, _pClaims0);
        address signerAccount1 = AccusationsLibrary.recoverMadNetSigner(_signature1, _pClaims1);

        require(
            signerAccount0 == signerAccount1,
            "Accusations: the signers of the proposals should be the same"
        );

        // ensure the hashes of blob0/1 are different
        require(
            keccak256(_pClaims0) != keccak256(_pClaims1),
            "Accusations: the PClaims are equal!"
        );

        PClaimsParserLibrary.PClaims memory pClaims0 = PClaimsParserLibrary.extractPClaims(
            _pClaims0
        );
        PClaimsParserLibrary.PClaims memory pClaims1 = PClaimsParserLibrary.extractPClaims(
            _pClaims1
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

        // ensure both accounts are applicable to a currently locked validator - Note<may be done in different layer?>
        require(
            IValidatorPool(_validatorPoolAddress()).isAccusable(signerAccount0),
            "Accusations: the signer of these proposals is not a valid validator!"
        );

        return signerAccount0;
    }
}
