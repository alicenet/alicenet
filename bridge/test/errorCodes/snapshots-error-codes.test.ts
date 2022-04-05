import { Contract } from "ethers";
import { assertConstantReturnsCorrectErrorCode, deployLibrary } from "./setup";

describe("Testing Snapshot error codes", async () => {
  let snapshotErrorCodesContract: Contract;

  beforeEach(async function () {
    snapshotErrorCodesContract = await deployLibrary("SnapshotsErrorCodes");
  });

  it("SNAPSHOT_ONLY_VALIDATORS_ALLOWED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      snapshotErrorCodesContract.SNAPSHOT_ONLY_VALIDATORS_ALLOWED,
      400
    );
  });
  it("SNAPSHOT_CONSENSUS_RUNNING returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      snapshotErrorCodesContract.SNAPSHOT_CONSENSUS_RUNNING,
      401
    );
  });
  it("SNAPSHOT_MIN_BLOCKS_INTERVAL_NOT_PASSED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      snapshotErrorCodesContract.SNAPSHOT_MIN_BLOCKS_INTERVAL_NOT_PASSED,
      402
    );
  });
  it("SNAPSHOT_CALLER_NOT_ETHDKG_PARTICIPANT returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      snapshotErrorCodesContract.SNAPSHOT_CALLER_NOT_ETHDKG_PARTICIPANT,
      403
    );
  });
  it("SNAPSHOT_WRONG_MASTER_PUBLIC_KEY returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      snapshotErrorCodesContract.SNAPSHOT_WRONG_MASTER_PUBLIC_KEY,
      404
    );
  });
  it("SNAPSHOT_SIGNATURE_VERIFICATION_FAILED returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      snapshotErrorCodesContract.SNAPSHOT_SIGNATURE_VERIFICATION_FAILED,
      405
    );
  });
  it("SNAPSHOT_INCORRECT_BLOCK_HEIGHT returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      snapshotErrorCodesContract.SNAPSHOT_INCORRECT_BLOCK_HEIGHT,
      406
    );
  });
  it("SNAPSHOT_INCORRECT_CHAIN_ID returns correct code", async () => {
    await assertConstantReturnsCorrectErrorCode(
      snapshotErrorCodesContract.SNAPSHOT_INCORRECT_CHAIN_ID,
      407
    );
  });
});
