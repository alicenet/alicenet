import { BigNumberish, ethers } from "ethers";
import {
  signedData,
  validatorsSnapshots,
} from "../math/assets/4-validators-1000-snapshots";
import {
  expect,
  factoryCallAny,
  factoryCallAnyFixture,
  Fixture,
  getFixture,
  getValidatorEthAccount,
  mineBlocks,
} from "../setup";
import {
  validatorsSnapshots as validatorsSnapshots2,
  validSnapshot1024,
  validSnapshot2048,
} from "../snapshots/assets/4-validators-snapshots-1";
import { createValidators, stakeValidators } from "../validatorPool/setup";
import { completeETHDKGRound } from "./setup";

describe("Ethdkg: Migrate state", () => {
  let fixture: Fixture;
  let validatorsAddress: string[];
  let validatorsShares: [
    BigNumberish,
    BigNumberish,
    BigNumberish,
    BigNumberish
  ][];

  beforeEach(async function () {
    fixture = await getFixture();
    validatorsAddress = [];
    validatorsShares = [];
    for (let i = 0; i < validatorsSnapshots.length; i++) {
      validatorsAddress.push(validatorsSnapshots[i].address);
      validatorsShares.push(validatorsSnapshots[i].gpkj);
    }
  });

  it("Should not be to do a migration of validators if not factory", async function () {
    await expect(
      fixture.ethdkg.migrateValidators(
        validatorsAddress,
        [1, 2, 3, 4],
        validatorsShares,
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk
      )
    ).to.be.revertedWith("2000");
  });

  it("Should not be to do a migration with mismatch state length", async function () {
    await expect(
      factoryCallAny(fixture.factory, fixture.ethdkg, "migrateValidators", [
        validatorsAddress,
        [1, 2, 3],
        validatorsShares,
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ])
    ).to.be.revertedWith("152");
    await expect(
      factoryCallAny(fixture.factory, fixture.ethdkg, "migrateValidators", [
        validatorsAddress,
        [1, 2, 3, 4],
        validatorsShares.slice(0, 3),
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ])
    ).to.be.revertedWith("152");
    await expect(
      factoryCallAny(fixture.factory, fixture.ethdkg, "migrateValidators", [
        validatorsAddress.slice(0, 3),
        [1, 2, 3, 4],
        validatorsShares,
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ])
    ).to.be.revertedWith("152");
  });

  it("Factory should be able to migrate validators", async function () {
    const receipt = await factoryCallAny(
      fixture.factory,
      fixture.ethdkg,
      "migrateValidators",
      [
        validatorsAddress,
        [1, 2, 3, 4],
        validatorsShares,
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ]
    );
    expect(receipt.status).to.be.equals(1);
  });

  it("Should not be able to run ethdkg after migration without scheduling maintenance", async function () {
    const validators = await createValidators(fixture, validatorsSnapshots);
    const stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );

    const receipt = await factoryCallAny(
      fixture.factory,
      fixture.ethdkg,
      "migrateValidators",
      [
        validatorsAddress,
        [1, 2, 3, 4],
        validatorsShares,
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ]
    );
    expect(receipt.status).to.be.equals(1);

    await expect(
      fixture.factory.callAny(
        fixture.validatorPool.address,
        0,
        fixture.validatorPool.interface.encodeFunctionData("initializeETHDKG")
      )
    ).to.be.revertedWith("801");
  });

  it("Should not be able to run more than 1 migration", async function () {
    const validators = await createValidators(fixture, validatorsSnapshots);
    const stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );

    const receipt = await factoryCallAny(
      fixture.factory,
      fixture.ethdkg,
      "migrateValidators",
      [
        validatorsAddress,
        [1, 2, 3, 4],
        validatorsShares,
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ]
    );
    expect(receipt.status).to.be.equals(1);

    await expect(
      fixture.factory.callAny(
        fixture.ethdkg.address,
        0,
        fixture.ethdkg.interface.encodeFunctionData("migrateValidators", [
          validatorsAddress,
          [1, 2, 3, 4],
          validatorsShares,
          validatorsSnapshots.length,
          0,
          0,
          100,
          validatorsSnapshots[0].mpk,
        ])
      )
    ).to.be.revertedWith("151");
  });

  it("Change validators after migration with scheduling maintenance + snapshots", async function () {
    const validators = await createValidators(fixture, validatorsSnapshots2);
    const stakingTokenIds = await stakeValidators(fixture, validators);
    validatorsAddress = [];
    validatorsShares = [];
    for (let i = 0; i < validatorsSnapshots.length; i++) {
      validatorsAddress.push(validatorsSnapshots2[i].address);
      validatorsShares.push(validatorsSnapshots2[i].gpkj);
    }
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );

    let receipt = await factoryCallAny(
      fixture.factory,
      fixture.ethdkg,
      "migrateValidators",
      [
        validatorsAddress,
        [1, 2, 3, 4],
        validatorsShares,
        validatorsSnapshots2.length,
        0,
        0,
        100,
        validatorsSnapshots2[0].mpk,
      ]
    );
    expect(receipt.status).to.be.equals(1);

    // migrating snapshots as well
    receipt = await factoryCallAny(
      fixture.factory,
      fixture.snapshots,
      "migrateSnapshots",
      [[validSnapshot1024.GroupSignature], [validSnapshot1024.BClaims]]
    );
    expect(receipt.status).to.be.equals(1);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "scheduleMaintenance"
    );
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
    let expectedChainId = 1;
    let expectedEpoch = 2;
    let expectedHeight = 2048;
    let expectedSafeToProceedConsensus = false;

    await expect(
      fixture.snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots2[1]))
        .snapshot(validSnapshot2048.GroupSignature, validSnapshot2048.BClaims)
    )
      .to.emit(fixture.snapshots, `SnapshotTaken`)
      .withArgs(
        expectedChainId,
        expectedEpoch,
        expectedHeight,
        ethers.utils.getAddress(validatorsSnapshots2[1].address),
        expectedSafeToProceedConsensus,
        validSnapshot2048.GroupSignature
      );

    // unregister validators and register news
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "unregisterValidators",
      [validators]
    );

    // registering the new validators
    const newValidators = await createValidators(fixture, validatorsSnapshots);
    const newStakingTokenIds = await stakeValidators(fixture, newValidators);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [newValidators, newStakingTokenIds]
    );
    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(
      validatorsSnapshots,
      {
        ethdkg: fixture.ethdkg,
        validatorPool: fixture.validatorPool,
      },
      expectedEpoch,
      expectedHeight,
      (
        await fixture.snapshots.getCommittedHeightFromLatestSnapshot()
      ).toNumber()
    );
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
    expectedChainId = 1;
    expectedEpoch = 3;
    expectedHeight = 3072;
    expectedSafeToProceedConsensus = true;
    await expect(
      fixture.snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots[1]))
        .snapshot(signedData[2].GroupSignature, signedData[2].BClaims)
    )
      .to.emit(fixture.snapshots, `SnapshotTaken`)
      .withArgs(
        expectedChainId,
        expectedEpoch,
        expectedHeight,
        ethers.utils.getAddress(validatorsSnapshots[1].address),
        expectedSafeToProceedConsensus,
        signedData[2].GroupSignature
      );
  });

  it("Run ethdkg with same validators after migration with scheduling maintenance + snapshots", async function () {
    const validators = await createValidators(fixture, validatorsSnapshots);
    const stakingTokenIds = await stakeValidators(fixture, validators);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "registerValidators",
      [validators, stakingTokenIds]
    );

    let receipt = await factoryCallAny(
      fixture.factory,
      fixture.ethdkg,
      "migrateValidators",
      [
        validatorsAddress,
        [1, 2, 3, 4],
        validatorsShares,
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ]
    );
    expect(receipt.status).to.be.equals(1);

    // migrating snapshots as well
    receipt = await factoryCallAny(
      fixture.factory,
      fixture.snapshots,
      "migrateSnapshots",
      [[signedData[0].GroupSignature], [signedData[0].BClaims]]
    );
    expect(receipt.status).to.be.equals(1);
    await factoryCallAnyFixture(
      fixture,
      "validatorPool",
      "scheduleMaintenance"
    );
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
    let expectedChainId = 1;
    let expectedEpoch = 2;
    let expectedHeight = 2048;
    let expectedSafeToProceedConsensus = false;

    await expect(
      fixture.snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots[1]))
        .snapshot(signedData[1].GroupSignature, signedData[1].BClaims)
    )
      .to.emit(fixture.snapshots, `SnapshotTaken`)
      .withArgs(
        expectedChainId,
        expectedEpoch,
        expectedHeight,
        ethers.utils.getAddress(validatorsSnapshots[1].address),
        expectedSafeToProceedConsensus,
        signedData[1].GroupSignature
      );

    await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    await completeETHDKGRound(
      validatorsSnapshots,
      {
        ethdkg: fixture.ethdkg,
        validatorPool: fixture.validatorPool,
      },
      expectedEpoch,
      expectedHeight,
      (
        await fixture.snapshots.getCommittedHeightFromLatestSnapshot()
      ).toNumber()
    );
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
    expectedChainId = 1;
    expectedEpoch = 3;
    expectedHeight = 3072;
    expectedSafeToProceedConsensus = true;
    await expect(
      fixture.snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots[1]))
        .snapshot(signedData[2].GroupSignature, signedData[2].BClaims)
    )
      .to.emit(fixture.snapshots, `SnapshotTaken`)
      .withArgs(
        expectedChainId,
        expectedEpoch,
        expectedHeight,
        ethers.utils.getAddress(validatorsSnapshots[1].address),
        expectedSafeToProceedConsensus,
        signedData[2].GroupSignature
      );
  });
});
