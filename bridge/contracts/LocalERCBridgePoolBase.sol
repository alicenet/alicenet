// SPDX-License-Identifier: MIT-open-group
pragma solidity ^0.8.16;
import "contracts/interfaces/IBridgePool.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errors/LocalERCBridgePoolBaseErrors.sol";
import "contracts/utils/AccusationsLibrary.sol";
import "contracts/libraries/errors/AccusationsErrors.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/libraries/parsers/TXInPreImageParserLibrary.sol";
import "contracts/libraries/parsers/PClaimsParserLibrary.sol";
import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/Snapshots.sol";

import "hardhat/console.sol";

/// @custom:salt LocalERCBridgePoolBase
/// @custom:deploy-type deployStatic
abstract contract LocalERCBridgePoolBase is IBridgePool, ImmutableSnapshots {
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    mapping (bytes32 => bool) consumedUTXOIDs;
    struct UTXO {
        uint32 chainID;
        address owner;
        uint256 tokenId;
        uint256 tokenAmount;
        uint256 fee;
        bytes32 txHash;
    }

    struct DepositParameters {
        uint256 tokenId;
        uint256 tokenAmount;
    }

    constructor() ImmutableFactory(msg.sender) {}

    /// @notice Transfer ERC20 or ERC721 tokens from sender for minting them on sidechain
    /// @param msgSender The address of ERC sender
    /// @param depositParameters encoded deposit parameters (ERC20:tokenAmount, ERC721:tokenId or ERC1155:tokenAmount+tokenId)
    function deposit(address msgSender, bytes calldata depositParameters) public virtual {}

    /// @notice This function validates an accusation of non-existent utxo consumption, as well as invalid deposit consumption.
    /// @param _bClaims the BClaims of the withdraw
    /// @param _bClaimsSigGroup the signature group of BClaims
    /// @param _txInPreImage the TXInPreImage consuming the invalid transaction
    /// @param _proofs an array of merkle proof structs in the following order:
    /// proof against StateRoot: Proof of inclusion or exclusion of the deposit or UTXO in the stateTrie
    /// proof of inclusion in TXRoot: Proof of inclusion of the transaction that included the invalid input in the txRoot trie.
    /// proof of inclusion in TXHash: Proof of inclusion of the invalid input (txIn) in the txHash trie (transaction tested against the TxRoot).
    function withdraw(
        bytes memory _bClaims,
        bytes memory _bClaimsSigGroup,
        bytes memory _txInPreImage,
        bytes[4] memory _proofs
    ) public virtual {
        // Require that the previous block is signed by correct group key for validator set.
        // _verifySignatureGroup(_bClaims, _bClaimsSigGroup);

        // Require that height delta is 1.
        BClaimsParserLibrary.BClaims memory bClaims = BClaimsParserLibrary.extractBClaims(_bClaims);
        console.logBytes32(bClaims.txRoot);

        // Require that chainID is equal.
        if (
            bClaims.chainId != ISnapshots(_snapshotsAddress()).getChainId()
        ) {
            revert AccusationsErrors.ChainIdDoesNotMatch(
                bClaims.chainId,
                bClaims.chainId,
                ISnapshots(_snapshotsAddress()).getChainId()
            );
        }


        // Validate ProofInclusionTxRoot against BClaims.HeaderRoot.
        MerkleProofParserLibrary.MerkleProof memory proofInclusionHeaderRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[3]);
        MerkleProofLibrary.verifyInclusion(proofInclusionHeaderRoot, bClaims.headerRoot);

        // Validate ProofInclusionTxRoot against BClaims.TxRoot.
        MerkleProofParserLibrary.MerkleProof memory proofInclusionTxRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[1]);
        MerkleProofLibrary.verifyInclusion(proofInclusionTxRoot, bClaims.txRoot);

        // Validate ProofOfInclusionTxHash against the target hash from ProofInclusionTxRoot.
        MerkleProofParserLibrary.MerkleProof
            memory proofOfInclusionTxHash = MerkleProofParserLibrary.extractMerkleProof(_proofs[2]);
        MerkleProofLibrary.verifyInclusion(proofOfInclusionTxHash, proofInclusionTxRoot.key);

        MerkleProofParserLibrary.MerkleProof memory proofAgainstStateRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[0]);

        if (proofAgainstStateRoot.key != proofOfInclusionTxHash.key) {
            revert AccusationsErrors.UTXODoesnotMatch(
                proofAgainstStateRoot.key,
                proofOfInclusionTxHash.key
            );
        }

        TXInPreImageParserLibrary.TXInPreImage memory txInPreImage = TXInPreImageParserLibrary
            .extractTXInPreImage(_txInPreImage);

                bytes32 computedUTXOID = AccusationsLibrary.computeUTXOID(
                    txInPreImage.consumedTxHash,
                    txInPreImage.consumedTxIdx
                );
                if (computedUTXOID != proofAgainstStateRoot.key) {
                    revert AccusationsErrors.MerkleProofKeyDoesNotMatchUTXOIDBeingSpent(
                        computedUTXOID,
                        proofAgainstStateRoot.key
                    );
                }
                if (consumedUTXOIDs[computedUTXOID] == true) {
                    revert LocalERCBridgePoolBaseErrors.UTXOAlreadyConsumed(
                        computedUTXOID
                    );
                }
        MerkleProofLibrary.verifyInclusion(proofAgainstStateRoot, bClaims.stateRoot);
        consumedUTXOIDs[computedUTXOID] = true;

    }


/*     /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
    /// @param _bClaims informed bClaims
    /// @param _bClaimsSigGroup bClaims signature
    /// @param _txInPreImage withdraw (burn) tx
    /// @param _proofs inclusion proofs for StateRoot, TxRoot, TxHash and HeaderRoot
    function withdraw(
        bytes memory _pClaims,
        bytes memory _pClaimsSig,
        bytes memory _bClaims,
        bytes memory _bClaimsSigGroup,
        bytes memory _txInPreImage,
        bytes[4] memory _proofs
    )
        public
        virtual
    {
        // _verifySignatureGroup(_bClaims, _bClaimsSigGroup);

        PClaimsParserLibrary.PClaims memory pClaims = PClaimsParserLibrary.extractPClaims(_pClaims);
/*         BClaimsParserLibrary.BClaims memory bClaims = Snapshots(_snapshotsAddress())
            .getBlockClaimsFromLatestSnapshot(); 

        // Validate ProofInclustionHeaderRoot against BClaims.TxRoot.
        MerkleProofParserLibrary.MerkleProof memory proofInclustionHeaderRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[3]);
        proofInclustionHeaderRoot.verifyInclusion(bClaims.headerRoot);
      
        // Require that chainID is equal.
        if (
            bClaims.chainId != ISnapshots(_snapshotsAddress()).getChainId()
        ) {
            revert LocalERCBridgePoolBaseErrors.ChainIdDoesNotMatch(
                bClaims.chainId,
                ISnapshots(_snapshotsAddress()).getChainId()
            );
        }


        // Validate ProofInclusionTxRoot against BClaims.TxRoot.
                MerkleProofParserLibrary.MerkleProof memory proofInclusionTxRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[1]);
        MerkleProofLibrary.verifyInclusion(proofInclusionTxRoot, pClaims.bClaims.txRoot);


        // Validate ProofOfInclusionTxHash against the target hash from ProofInclusionTxRoot.
        MerkleProofParserLibrary.MerkleProof
            memory proofOfInclusionTxHash = MerkleProofParserLibrary.extractMerkleProof(_proofs[2]);
        MerkleProofLibrary.verifyInclusion(proofOfInclusionTxHash, proofInclusionTxRoot.key);

        MerkleProofParserLibrary.MerkleProof memory proofAgainstStateRoot = MerkleProofParserLibrary
            .extractMerkleProof(_proofs[0]);
        if (proofAgainstStateRoot.key != proofOfInclusionTxHash.key) {
            revert LocalERCBridgePoolBaseErrors.UTXODoesnotMatch(
                proofAgainstStateRoot.key,
                proofOfInclusionTxHash.key
            );
        }

        TXInPreImageParserLibrary.TXInPreImage memory txInPreImage = TXInPreImageParserLibrary
            .extractTXInPreImage(_txInPreImage);

 *//*         // checking if we are consuming a deposit or an UTXO
        if (txInPreImage.consumedTxIdx == 0xFFFFFFFF) {
            // Double spending problem, i.e, consuming a deposit that was already consumed
            if (txInPreImage.consumedTxHash != proofAgainstStateRoot.key) {
                revert AccusationsErrors.MerkleProofKeyDoesNotMatchConsumedDepositKey(
                    txInPreImage.consumedTxHash,
                    proofAgainstStateRoot.key
                );
            }
            MerkleProofLibrary.verifyInclusion(proofAgainstStateRoot, bClaims.stateRoot);
            // todo: deposit that doesn't exist in the chain. Maybe split this in separate functions?
        } else {
            // Consuming a non existing UTXO
            {
                bytes32 computedUTXOID = AccusationsLibrary.computeUTXOID(
                    txInPreImage.consumedTxHash,
                    txInPreImage.consumedTxIdx
                );
                if (computedUTXOID != proofAgainstStateRoot.key) {
                    revert AccusationsErrors.MerkleProofKeyDoesNotMatchUTXOIDBeingSpent(
                        computedUTXOID,
                        proofAgainstStateRoot.key
                    );
                }
            }
            MerkleProofLibrary.verifyNonInclusion(proofAgainstStateRoot, bClaims.stateRoot);
        } */



/*         if (proofAgainstStateRoot.key != proofOfInclusionTxHash.key) {
            revert AccusationsErrors.UTXODoesnotMatch(
                proofAgainstStateRoot.key,
                proofOfInclusionTxHash.key
            );
        } */
/*         UTXO memory burnedUTXO = abi.decode(encodedBurnedUTXO, (UTXO));
        if (burnedUTXO.owner != msg.sender) {
            revert LocalERCBridgePoolBaseErrors.ReceiverIsNotOwnerOnProofOfBurnUTXO();
        }
 
         UTXO memory burnedUTXO = abi.decode(encodedBurnedUTXO, (UTXO));
        if (burnedUTXO.owner != msg.sender) {
            revert LocalERCBridgePoolBaseErrors.ReceiverIsNotOwnerOnProofOfBurnUTXO();
        } 
    }
 */
    /// @notice This function verifies the signature group of a BClaims.
    /// @param _bClaims the BClaims of the accusation
    /// @param _bClaimsSigGroup the signature group of Pclaims
    function _verifySignatureGroup(bytes memory _bClaims, bytes memory _bClaimsSigGroup)
        internal
        view
    {
        uint256[4] memory publicKey;
        uint256[2] memory signature;
        (publicKey, signature) = RCertParserLibrary.extractSigGroup(_bClaimsSigGroup, 0);

        // todo: check if the signature is equals to any of the previous master public key?

        if (
            !CryptoLibrary.verifySignatureASM(
                abi.encodePacked(keccak256(_bClaims)),
                signature,
                publicKey
            )
        ) {
            revert LocalERCBridgePoolBaseErrors.SignatureVerificationFailed();
        }
    }


}
