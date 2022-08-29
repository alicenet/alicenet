import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumberish } from "ethers";
import { ethers } from "hardhat";
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
  let admin: SignerWithAddress;
  let validatorsAddress: string[];
  let validatorsShares: [
    BigNumberish,
    BigNumberish,
    BigNumberish,
    BigNumberish
  ][];

  beforeEach(async function () {
    fixture = await getFixture();
    [admin] = await ethers.getSigners();
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
    )
      .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
      .withArgs(admin.address, fixture.factory.address);
  });

  it("Should not be to do a migration with mismatch state length", async function () {
    const validatorIndexes = [1, 2, 3];
    await expect(
      factoryCallAny(fixture.factory, fixture.ethdkg, "migrateValidators", [
        validatorsAddress,
        validatorIndexes,
        validatorsShares,
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ])
    )
      .to.be.revertedWithCustomError(
        fixture.ethdkg,
        "MigrationInputDataMismatch"
      )
      .withArgs(
        validatorsAddress.length,
        validatorIndexes.length,
        validatorsShares.length
      );
    const correctValidatorIndexes = [1, 2, 3, 4];
    await expect(
      factoryCallAny(fixture.factory, fixture.ethdkg, "migrateValidators", [
        validatorsAddress,
        correctValidatorIndexes,
        validatorsShares.slice(0, 3),
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ])
    )
      .to.be.revertedWithCustomError(
        fixture.ethdkg,
        "MigrationInputDataMismatch"
      )
      .withArgs(validatorsAddress.length, correctValidatorIndexes.length, 3);
    await expect(
      factoryCallAny(fixture.factory, fixture.ethdkg, "migrateValidators", [
        validatorsAddress.slice(0, 3),
        correctValidatorIndexes,
        validatorsShares,
        validatorsSnapshots.length,
        0,
        0,
        100,
        validatorsSnapshots[0].mpk,
      ])
    )
      .to.be.revertedWithCustomError(
        fixture.ethdkg,
        "MigrationInputDataMismatch"
      )
      .withArgs(3, correctValidatorIndexes.length, validatorsShares.length);
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
    ).to.be.revertedWithCustomError(fixture.validatorPool, "ConsensusRunning");
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
    )
      .to.be.revertedWithCustomError(
        fixture.ethdkg,
        "MigrationRequiresZeroNonce"
      )
      .withArgs(1);
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
        validSnapshot2048.GroupSignatureDeserialized?.[0],
        validSnapshot2048.GroupSignatureDeserialized?.[1],
        validSnapshot2048.BClaimsDeserialized
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
        [
          "0x0062cd4187d44be6f7977e5cbfc18066c3d5029bc6ab1e0ae5b1dd20a691fc6d",
          "0x08648a63a6690c930265e93c86ec421d6a7ca06504c6b9509640cbd794a1459a",
          "0x0a0837516f6bdc0ff9fd69776b2d7928432958b31551d10e921cc261f290b23c",
          "0x06ce5812bf9f76f2dc04d272dd2e0ff8d2424d1e9f19c22da1ad5d2294463428",
        ],
        [
          "0x0413ba20de19c18c75e268419c3f6fd86cc4d9a75e3b1127ce180ed93b95bff6",
          "0x09401a3cf161c59e5e34e0da23593f556c9f834e90264a474933fd8235202b0c",
        ],
        [
          expectedChainId,
          expectedHeight,
          0,
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        ]
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
        [
          "0x0062cd4187d44be6f7977e5cbfc18066c3d5029bc6ab1e0ae5b1dd20a691fc6d",
          "0x08648a63a6690c930265e93c86ec421d6a7ca06504c6b9509640cbd794a1459a",
          "0x0a0837516f6bdc0ff9fd69776b2d7928432958b31551d10e921cc261f290b23c",
          "0x06ce5812bf9f76f2dc04d272dd2e0ff8d2424d1e9f19c22da1ad5d2294463428",
        ],
        [
          "0x2a060cf960125afbb506d17644798a7e5c877e9e82b6bf426b7287d543ed59e0",
          "0x09d52226e77e1bf3321bff5320989ee80f27a564d25bc4daaf4bb0640c73d7b0",
        ],
        [
          expectedChainId,
          expectedHeight,
          0,
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        ]
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
        [
          "0x0062cd4187d44be6f7977e5cbfc18066c3d5029bc6ab1e0ae5b1dd20a691fc6d",
          "0x08648a63a6690c930265e93c86ec421d6a7ca06504c6b9509640cbd794a1459a",
          "0x0a0837516f6bdc0ff9fd69776b2d7928432958b31551d10e921cc261f290b23c",
          "0x06ce5812bf9f76f2dc04d272dd2e0ff8d2424d1e9f19c22da1ad5d2294463428",
        ],
        [
          "0x0413ba20de19c18c75e268419c3f6fd86cc4d9a75e3b1127ce180ed93b95bff6",
          "0x09401a3cf161c59e5e34e0da23593f556c9f834e90264a474933fd8235202b0c",
        ],
        [
          expectedChainId,
          expectedHeight,
          0,
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        ]
      );
  });
});
