import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing MerkleProofParserLibrary error codes", async () => {
  let merkleProofParserLibErrorCodesContract: Contract;

  beforeEach(async function () {
    merkleProofParserLibErrorCodesContract = await deployLibrary(
      "MerkleProofParserLibraryErrorCodes"
    );
  });

  it("MERKLEPROOFPARSERLIB_INVALID_PROOF_MINIMUM_SIZE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      merkleProofParserLibErrorCodesContract.MERKLEPROOFPARSERLIB_INVALID_PROOF_MINIMUM_SIZE,
      1200
    );
  });
  it("MERKLEPROOFPARSERLIB_INVALID_PROOF_SIZE returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      merkleProofParserLibErrorCodesContract.MERKLEPROOFPARSERLIB_INVALID_PROOF_SIZE,
      1201
    );
  });
  it("MERKLEPROOFPARSERLIB_INVALID_KEY_HEIGHT returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      merkleProofParserLibErrorCodesContract.MERKLEPROOFPARSERLIB_INVALID_KEY_HEIGHT,
      1202
    );
  });
});
