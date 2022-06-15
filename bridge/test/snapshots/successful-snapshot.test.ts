import { ethers } from "hardhat";
import { Snapshots } from "../../typechain-types";
import { expect } from "../chai-setup";
import { completeETHDKGRound } from "../ethdkg/setup";
import {
  Fixture,
  getFixture,
  getValidatorEthAccount,
  mineBlocks,
} from "../setup";
import {
  invalidSnapshot500,
  invalidSnapshotChainID2,
  invalidSnapshotIncorrectSig,
  validatorsSnapshots as validatorsSnapshots1,
  validSnapshot1024,
} from "./assets/4-validators-snapshots-1";

describe("Snapshots: With successful ETHDKG round completed", () => {
  let fixture: Fixture;
  let snapshots: Snapshots;
  beforeEach(async function () {
    fixture = await getFixture(true, false);

    await completeETHDKGRound(validatorsSnapshots1, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });

    snapshots = fixture.snapshots as Snapshots;
  });

  it("Reverts when validator not elected to do snapshot", async function () {
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
    const junkData =
      "0x0000000000000000000000000000000000000000000000000000006d6168616d";
    await expect(
      snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots1[0]))
        .snapshot(junkData, junkData)
    ).to.be.revertedWith("1401");
  });

  it("Reverts when snapshot state contains invalid height", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots1[invalidSnapshot500.validatorIndex]
          )
        )
        .snapshot(invalidSnapshot500.GroupSignature, invalidSnapshot500.BClaims)
    ).to.be.revertedWith("406");
  });

  it("Reverts when snapshot state contains invalid chain id", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots1[invalidSnapshotChainID2.validatorIndex]
          )
        )
        .snapshot(
          invalidSnapshotChainID2.GroupSignature,
          invalidSnapshotChainID2.BClaims
        )
    ).to.be.revertedWith("407");
  });

  // todo wrong public key failure happens first with this state
  it("Reverts when snapshot state contains incorrect signature", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots1[invalidSnapshotIncorrectSig.validatorIndex]
          )
        )
        .snapshot(
          validSnapshot1024.GroupSignature,
          invalidSnapshotIncorrectSig.BClaims
        )
    ).to.be.revertedWith("405");
  });

  it("Reverts when snapshot state contains incorrect public key", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots1[invalidSnapshotIncorrectSig.validatorIndex]
          )
        )
        .snapshot(
          invalidSnapshotIncorrectSig.GroupSignature,
          invalidSnapshotIncorrectSig.BClaims
        )
    ).to.be.revertedWith("404");
  });

  it("Successfully performs snapshot", async function () {
    const expectedChainId = 1;
    const expectedEpoch = 1;
    const expectedHeight = validSnapshot1024.height;
    const expectedSafeToProceedConsensus = true;
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
    await expect(
      snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots1[0]))
        .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims)
    )
      .to.emit(snapshots, `SnapshotTaken`)
      .withArgs(
        expectedChainId,
        expectedEpoch,
        expectedHeight,
        ethers.utils.getAddress(validatorsSnapshots1[0].address),
        expectedSafeToProceedConsensus,
        validSnapshot1024.GroupSignature
      );
  });

  /*
  FYI this scenario is not possible to cover due to the fact that no validators can be registered but not participate in the ETHDKG round.

  it('Does not allow snapshot caller did not participate in the last ETHDKG round', async function () {
    await expect(
      snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots1[0]))
        .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims)
    ).to.be.revertedWith(
      `Snapshots: Caller didn't participate in the last ethdkg round!`
    )
  }) */
});
