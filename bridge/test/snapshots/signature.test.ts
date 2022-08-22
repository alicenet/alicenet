import { Snapshots } from "../../typechain-types";
import { expect } from "../chai-setup";
import { completeETHDKGRound } from "../ethdkg/setup";
import { Fixture, getFixture, getValidatorEthAccount } from "../setup";
import {
  invalidSnapshotIncorrectSig,
  validatorsSnapshots as validatorsSnapshots1,
  validSnapshot1024,
} from "./assets/4-validators-snapshots-1";

describe("Snapshots: signature check algorithm", () => {
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

  it("Correctly check honest data", async function () {
    expect(
      await snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots1[invalidSnapshotIncorrectSig.validatorIndex]
          )
        )
        .checkBClaimsSignature(
          validSnapshot1024.GroupSignature,
          validSnapshot1024.BClaims
        )
    ).to.be.equals(true);
  });

  it("Reverts when tampered state", async function () {
    const tamperedBClaims = validSnapshot1024.BClaims + "7777";
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots1[invalidSnapshotIncorrectSig.validatorIndex]
          )
        )
        .checkBClaimsSignature(
          validSnapshot1024.GroupSignature,
          tamperedBClaims
        )
    ).to.be.revertedWithCustomError(snapshots, "SignatureVerificationFailed");
  });

  it("Reverts when state contains incorrect signature", async function () {
    await expect(
      snapshots
        .connect(
          await getValidatorEthAccount(
            validatorsSnapshots1[invalidSnapshotIncorrectSig.validatorIndex]
          )
        )
        .checkBClaimsSignature(
          validSnapshot1024.GroupSignature,
          invalidSnapshotIncorrectSig.BClaims
        )
    ).to.be.revertedWithCustomError(snapshots, "SignatureVerificationFailed");
  });

  it("Reverts when state contains incorrect public key", async function () {
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
        .checkBClaimsSignature(
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
});
