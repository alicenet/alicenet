import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
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
import { createValidators, stakeValidators } from "../validatorPool/setup";
import { invalidSnapshot500 } from "./assets/4-validators-snapshots-1";

describe("Snapshots: Migrate state", () => {
  let fixture: Fixture;
  let expectedBlockNumber: bigint;
  let admin: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getFixture();
    [admin] = fixture.namedSigners;
    const validators = await createValidators(fixture, validatorsSnapshots);
    const stakingTokenIds = await stakeValidators(fixture, validators);
    const validatorsShares = [];
    for (let i = 0; i < validatorsSnapshots.length; i++) {
      validatorsShares.push(validatorsSnapshots[i].gpkj);
    }
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
        validators,
        stakingTokenIds,
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

  it("Should not be to do a migration of snapshots if not factory", async function () {
    await expect(
      fixture.snapshots.migrateSnapshots(
        [signedData[0].GroupSignature],
        [signedData[0].BClaims]
      )
    )
      .to.be.revertedWithCustomError(fixture.bToken, `OnlyFactory`)
      .withArgs(admin.address, fixture.factory.address);
  });

  it("Should not be to do a migration with mismatch state length", async function () {
    await expect(
      factoryCallAny(fixture.factory, fixture.snapshots, "migrateSnapshots", [
        [signedData[0].GroupSignature],
        [signedData[0].BClaims, signedData[0].BClaims],
      ])
    )
      .to.be.revertedWithCustomError(
        fixture.snapshots,
        "MigrationInputDataMismatch"
      )
      .withArgs(1, 2);
  });

  it("Should not be to do a migration with incorrect block height", async function () {
    await expect(
      factoryCallAny(fixture.factory, fixture.snapshots, "migrateSnapshots", [
        [invalidSnapshot500.GroupSignature, signedData[0].GroupSignature],
        [invalidSnapshot500.BClaims, signedData[0].BClaims],
      ])
    )
      .to.be.revertedWithCustomError(fixture.snapshots, "InvalidBlockHeight")
      .withArgs(invalidSnapshot500.height);
    await expect(
      factoryCallAny(fixture.factory, fixture.snapshots, "migrateSnapshots", [
        [signedData[0].GroupSignature, invalidSnapshot500.GroupSignature],
        [signedData[0].BClaims, invalidSnapshot500.BClaims],
      ])
    )
      .to.be.revertedWithCustomError(fixture.snapshots, "InvalidBlockHeight")
      .withArgs(invalidSnapshot500.height);
    await expect(
      factoryCallAny(fixture.factory, fixture.snapshots, "migrateSnapshots", [
        [invalidSnapshot500.GroupSignature],
        [invalidSnapshot500.BClaims],
      ])
    )
      .to.be.revertedWithCustomError(fixture.snapshots, "InvalidBlockHeight")
      .withArgs(invalidSnapshot500.height);
  });

  it("Should be able to do snapshots after migration", async function () {
    const receipt = await factoryCallAny(
      fixture.factory,
      fixture.snapshots,
      "migrateSnapshots",
      [
        [
          signedData[500].GroupSignature,
          signedData[501].GroupSignature,
          signedData[502].GroupSignature,
        ],
        [
          signedData[500].BClaims,
          signedData[501].BClaims,
          signedData[502].BClaims,
        ],
      ]
    );
    expect(receipt.status).to.be.equals(1);
    expectedBlockNumber = BigInt(receipt.blockNumber);

    const expectedChainId = 1;
    const expectedEpoch = 504;
    const expectedHeight = 516096;
    const expectedSafeToProceedConsensus = true;
    await mineBlocks(
      (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    );
    await expect(
      fixture.snapshots
        .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
        .snapshot(signedData[503].GroupSignature, signedData[503].BClaims)
    )
      .to.emit(fixture.snapshots, `SnapshotTaken`)
      .withArgs(
        expectedChainId,
        expectedEpoch,
        expectedHeight,
        ethers.utils.getAddress(validatorsSnapshots[0].address),
        expectedSafeToProceedConsensus,
        [
          "0x0062cd4187d44be6f7977e5cbfc18066c3d5029bc6ab1e0ae5b1dd20a691fc6d",
          "0x08648a63a6690c930265e93c86ec421d6a7ca06504c6b9509640cbd794a1459a",
          "0x0a0837516f6bdc0ff9fd69776b2d7928432958b31551d10e921cc261f290b23c",
          "0x06ce5812bf9f76f2dc04d272dd2e0ff8d2424d1e9f19c22da1ad5d2294463428",
        ],
        [
          "0x21729f8d0a3ccfdd2751ecf67834e7247f23c3f3147c5278908758cb048d9780",
          "0x0d20314089ff3c8caf9d109659b4341630228dcc11412652374b6d94f61e1e11",
        ],
        [
          1,
          516096,
          0,
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
          "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470",
        ]
      );
  });

  describe("After a successful migration", () => {
    beforeEach(async function () {
      const receipt = await factoryCallAny(
        fixture.factory,
        fixture.snapshots,
        "migrateSnapshots",
        [
          [
            signedData[74].GroupSignature,
            signedData[75].GroupSignature,
            signedData[76].GroupSignature,
            signedData[77].GroupSignature,
            signedData[78].GroupSignature,
          ],
          [
            signedData[74].BClaims,
            signedData[75].BClaims,
            signedData[76].BClaims,
            signedData[77].BClaims,
            signedData[78].BClaims,
          ],
        ]
      );
      expect(receipt.status).to.be.equals(1);
      expectedBlockNumber = BigInt(receipt.blockNumber);
    });

    it("Factory should succeed doing a migration of snapshots", async function () {
      const expectedAliceNetHeights = [76800n, 77824n, 78848n, 79872n, 80896n];
      const expectedEpochs = [75n, 76n, 77n, 78n, 79n];

      expect((await fixture.snapshots.getEpoch()).toBigInt()).to.be.equal(
        expectedEpochs[expectedEpochs.length - 1]
      );
      expect(
        (
          await fixture.snapshots.getCommittedHeightFromLatestSnapshot()
        ).toBigInt()
      ).to.be.equal(expectedBlockNumber);
      expect(
        (
          await fixture.snapshots.getAliceNetHeightFromLatestSnapshot()
        ).toBigInt()
      ).to.be.equal(
        expectedAliceNetHeights[expectedAliceNetHeights.length - 1]
      );

      // making sure that all snapshots were written correctly
      for (let i = 0; i < expectedAliceNetHeights.length; i++) {
        expect(
          (
            await fixture.snapshots.getCommittedHeightFromSnapshot(
              expectedEpochs[i]
            )
          ).toBigInt()
        ).to.be.equal(expectedBlockNumber);
        expect(
          (
            await fixture.snapshots.getAliceNetHeightFromSnapshot(
              expectedEpochs[i]
            )
          ).toBigInt()
        ).to.be.equal(expectedAliceNetHeights[i]);
      }
    });

    it("Should not be able to do another migration after first migration", async function () {
      await expect(
        factoryCallAny(fixture.factory, fixture.snapshots, "migrateSnapshots", [
          [signedData[0].GroupSignature],
          [signedData[0].BClaims],
        ])
      ).to.be.revertedWithCustomError(
        fixture.snapshots,
        `MigrationNotAllowedAtCurrentEpoch`
      );
    });
  });
});
