import { ethers } from "hardhat";
import { Snapshots } from "../../typechain-types";
import { expect } from "../chai-setup";
import { completeETHDKGRound } from "../ethdkg/setup";
import {
  factoryCallAnyFixture,
  getFixture,
  getValidatorEthAccount,
  mineBlocks,
} from "../setup";
import { createValidators, stakeValidators } from "../validatorPool/setup";
import {
  validatorsSnapshots as validatorsSnapshots1,
  validSnapshot1024,
} from "./assets/4-validators-snapshots-1";
import {
  validatorsSnapshots as validatorsSnapshots2,
  validSnapshot2048,
} from "./assets/4-validators-snapshots-2";

describe("Snapshots: With successful ETHDKG round completed and validatorPool", () => {
  it("Successfully performs snapshot then change the validators and perform another snapshot", async function () {
    let expectedChainId = 1;
    let expectedEpoch = 1;
    let expectedHeight = validSnapshot1024.height as number;
    let expectedSafeToProceedConsensus = false;
    const fixture = await getFixture();
    const snapshots = fixture.snapshots as Snapshots;
    const validators = await createValidators(fixture, validatorsSnapshots1);
    const stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(validatorsSnapshots1, {
      ethdkg: fixture.ethdkg,
      validatorPool: fixture.validatorPool,
    });
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "scheduleMaintenance"
    );
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

    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterValidators",
      [validators]
    );

    // registering the new validators
    const newValidators = await createValidators(fixture, validatorsSnapshots2);
    const newStakingTokenIds = await stakeValidators(fixture, newValidators);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [newValidators, newStakingTokenIds]
    );
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(
      validatorsSnapshots2,
      {
        ethdkg: fixture.ethdkg,
        validatorPool: fixture.validatorPool,
      },
      expectedEpoch,
      expectedHeight,
      (await snapshots.getCommittedHeightFromLatestSnapshot()).toNumber()
    );
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
    expectedChainId = 1;
    expectedEpoch = 2;
    expectedHeight = validSnapshot2048.height as number;
    expectedSafeToProceedConsensus = true;
    await expect(
      snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots2[0]))
        .snapshot(validSnapshot2048.GroupSignature, validSnapshot2048.BClaims)
    )
      .to.emit(snapshots, `SnapshotTaken`)
      .withArgs(
        expectedChainId,
        expectedEpoch,
        expectedHeight,
        ethers.utils.getAddress(validatorsSnapshots2[0].address),
        expectedSafeToProceedConsensus,
        validSnapshot2048.GroupSignatureDeserialized?.[0],
        validSnapshot2048.GroupSignatureDeserialized?.[1],
        validSnapshot2048.BClaimsDeserialized
      );
  });
});
