// SPDX-License-Identifier: MIT-open-group 
pragma solidity ^0.8.16;
import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "contracts/interfaces/IERC20Transferable.sol";
import "contracts/utils/ImmutableAuth.sol";
import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import "contracts/libraries/errors/BridgePoolErrors.sol";
import "contracts/libraries/parsers/MerkleProofParserLibrary.sol";
import "contracts/utils/MerkleProofLibrary.sol";
import "contracts/Snapshots.sol";
import "contracts/libraries/parsers/BClaimsParserLibrary.sol";
import "contracts/utils/ERC20SafeTransfer.sol";

/// @custom:salt LocalERCBridgePoolBase
/// @custom:deploy-type deployStatic
abstract contract LocalERCBridgePoolBase is ImmutableSnapshots {
    using MerkleProofParserLibrary for bytes;
    using MerkleProofLibrary for MerkleProofParserLibrary.MerkleProof;

    struct UTXO {
        uint32 chainID;
        address owner;
        uint256 value;
        uint256 fee;
        bytes32 txHash;
    }

    constructor() ImmutableFactory(msg.sender) {}

    /// @notice Transfer ERC20 or ERC721 tokens from sender for minting them on sidechain
    /// @param msgSender The address of ERC sender
    /// @param number Token ammount if fungible or token id if non-fungible
    function deposit(address msgSender, uint256 number) public virtual {}

    /*     /// @notice Transfer tokens ERC1155 from sender for minting them on sidechain
    /// @param msgSender The address of ERC sender
    /// @param tokenId The token id on ERC1155
    /// @param number Token ammount if fungible or token id if non-fungible
    function deposit(address msgSender, uint256 tokenId, uint256 number) public virtual {
    }
 */
    /// @notice Transfer tokens to sender upon a verificable proof of burn in sidechain
    /// @param encodedMerkleProof The merkle proof
    /// @param encodedBurnedUTXO The burned UTXO
    function withdraw(bytes memory encodedMerkleProof, bytes memory encodedBurnedUTXO)
        public
        virtual
    {
        BClaimsParserLibrary.BClaims memory bClaims = Snapshots(_snapshotsAddress())
            .getBlockClaimsFromLatestSnapshot();
        MerkleProofParserLibrary.MerkleProof memory merkleProof = encodedMerkleProof
            .extractMerkleProof();
        UTXO memory burnedUTXO = abi.decode(encodedBurnedUTXO, (UTXO));
        if (burnedUTXO.owner != msg.sender) {
            revert BridgePoolErrors.ReceiverIsNotOwnerOnProofOfBurnUTXO();
        }
        merkleProof.verifyInclusion(bClaims.stateRoot);
    }
}
