import { ethers } from "hardhat";
import { Snapshots } from "../../typechain-types";
import { expect } from "../chai-setup";
import { completeETHDKGRound } from "../ethdkg/setup";
import {
  factoryCallAnyFixture,
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
  validSnapshot2048,
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
    await factoryCallAnyFixture(
      fixture,
      "snapshots",
      "setSnapshotDesperationDelay",
      [30000000n]
    );
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
    await expect(
      snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots1[2]))
        .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims)
    )
      .to.be.revertedWithCustomError(snapshots, "ValidatorNotElected")
      .withArgs(
        2,
        3,
        0,
        "0xe49578123478a2663c878cc45bdda03dd0ffa36f7f8f6517d849e726450aa04f"
      );
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
    )
      .to.be.revertedWithCustomError(fixture.snapshots, "InvalidBlockHeight")
      .withArgs(invalidSnapshot500.height);
  });

  it("Reverts when snapshot state is the future", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots1[validSnapshot2048.validatorIndex]
          )
        )
        .snapshot(validSnapshot2048.GroupSignature, validSnapshot2048.BClaims)
    )
      .to.be.revertedWithCustomError(fixture.snapshots, "InvalidBlockHeight")
      .withArgs(validSnapshot2048.height);
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
    )
      .to.be.revertedWithCustomError(fixture.snapshots, "InvalidChainId")
      .withArgs(2);
  });

  // todo wrong public key failure happens first with this state
  it("Reverts when snapshot state contains incorrect signature", async function () {
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
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
    ).to.be.revertedWithCustomError(snapshots, "SignatureVerificationFailed");
  });

  it("Reverts when snapshot state contains incorrect public key", async function () {
    const expectedCalculatedPublicKeyHash =
      "0x888ea4bcd71f772f1af058866a2234d1d1b0967c67a5b9d82248f8ad8d8c144c";
    const expectedMasterPublicKeyHash =
      "0x381f9c36df7c05b341eaf3708d6d05d9343cdcbccaf5989da9880024a9a8a4d7";
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
    )
      .to.be.revertedWithCustomError(
        fixture.snapshots,
        "InvalidMasterPublicKey"
      )
      .withArgs(expectedCalculatedPublicKeyHash, expectedMasterPublicKeyHash);
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
        validSnapshot1024.GroupSignatureDeserialized?.[0],
        validSnapshot1024.GroupSignatureDeserialized?.[1],
        validSnapshot1024.BClaimsDeserialized
      );
  });
});
