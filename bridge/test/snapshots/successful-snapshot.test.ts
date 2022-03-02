import { Fixture, getFixture, getValidatorEthAccount } from "../setup";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { completeETHDKGRound } from "../ethdkg/setup";
import { BigNumber } from "ethers";
import {
  validatorsSnapshots,
  validSnapshot1024,
  invalidSnapshot500,
  invalidSnapshotChainID2,
  invalidSnapshotIncorrectSig,
  validSnapshot2048,
} from "./assets/4-validators-snapshots-1";
import { Snapshots } from "../../typechain-types";

describe("Snapshots: With successful ETHDKG round completed", () => {
  let fixture: Fixture;
  let snapshots: Snapshots;
  beforeEach(async function () {
    fixture = await getFixture(true, false);

    await completeETHDKGRound(validatorsSnapshots, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });

    snapshots = fixture.snapshots as Snapshots;
  });

  it("Reverts when validator not elected to do snapshot", async function () {
    let junkData =
      "0x0000000000000000000000000000000000000000000000000000006d6168616d";
    await expect(
      snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
        .snapshot(junkData, junkData)
    ).to.be.revertedWith(`RCertParserLibrary: Not enough bytes to extract`);
  });

  it("Reverts when snapshot data contains invalid height", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots[invalidSnapshot500.validatorIndex]
          )
        )
        .snapshot(invalidSnapshot500.GroupSignature, invalidSnapshot500.BClaims)
    ).to.be.revertedWith(`Snapshots: Incorrect Madnet height for snapshot!`);
  });

  it("Reverts when snapshot data contains invalid chain id", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots[invalidSnapshotChainID2.validatorIndex]
          )
        )
        .snapshot(
          invalidSnapshotChainID2.GroupSignature,
          invalidSnapshotChainID2.BClaims
        )
    ).to.be.revertedWith(`Snapshots: Incorrect chainID for snapshot!`);
  });

  // todo wrong public key failure happens first with this data
  it("Reverts when snapshot data contains incorrect signature", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots[invalidSnapshotIncorrectSig.validatorIndex]
          )
        )
        .snapshot(
          validSnapshot1024.GroupSignature,
          invalidSnapshotIncorrectSig.BClaims
        )
    ).to.be.revertedWith(`Snapshots: Signature verification failed!`);
  });

  it("Reverts when snapshot data contains incorrect public key", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots[invalidSnapshotIncorrectSig.validatorIndex]
          )
        )
        .snapshot(
          invalidSnapshotIncorrectSig.GroupSignature,
          invalidSnapshotIncorrectSig.BClaims
        )
    ).to.be.revertedWith(`Snapshots: Wrong master public key!`);
  });

  it("Successfully performs snapshot", async function () {
    const expectedChainId = 1;
    const expectedEpoch = 1;
    const expectedHeight = validSnapshot1024.height;
    const expectedSafeToProceedConsensus = true;

    await expect(
      snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
        .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims)
    )
      .to.emit(snapshots, `SnapshotTaken`)
      .withArgs(
        expectedChainId,
        expectedEpoch,
        expectedHeight,
        ethers.utils.getAddress(validatorsSnapshots[0].address),
        expectedSafeToProceedConsensus,
        validSnapshot1024.GroupSignature
      );
  });

  /*
  FYI this scenario is not possible to cover due to the fact that no validators can be registered but not participate in the ETHDKG round.

  it('Does not allow snapshot caller did not participate in the last ETHDKG round', async function () {
    await expect(
      snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
        .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims)
    ).to.be.revertedWith(
      `Snapshots: Caller didn't participate in the last ethdkg round!`
    )
  })*/
});
