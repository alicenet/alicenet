import { BigNumberish } from "ethers";
import { validatorsSnapshots } from "../math/assets/4-validators-1000-snapshots";
import { expect, factoryCallAny, Fixture, getFixture } from "../setup";

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

  it("Should not be to do a migration with mismatch data length", async function () {
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
    ).to.be.revertedWith("409");
  });

  //   it("Should be able to do snapshots after migration", async function () {
  //     const validators = await createValidators(fixture, validatorsSnapshots);
  //     const stakingTokenIds = await stakeValidators(fixture, validators);
  //     await factoryCallAnyFixture(
  //       fixture,
  //       "validatorPool",
  //       "registerValidators",
  //       [validators, stakingTokenIds]
  //     );
  //     await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
  //     await completeETHDKGRound(validatorsSnapshots, {
  //       ethdkg: fixture.ethdkg,
  //       validatorPool: fixture.validatorPool,
  //     });

  //     const receipt = await factoryCallAny(
  //       fixture.factory,
  //       fixture.snapshots,
  //       "migrateSnapshots",
  //       [
  //         [
  //           signedData[500].GroupSignature,
  //           signedData[501].GroupSignature,
  //           signedData[502].GroupSignature,
  //         ],
  //         [
  //           signedData[500].BClaims,
  //           signedData[501].BClaims,
  //           signedData[502].BClaims,
  //         ],
  //       ]
  //     );
  //     expect(receipt.status).to.be.equals(1);
  //     expectedBlockNumber = BigInt(receipt.blockNumber);

  //     const expectedChainId = 1;
  //     const expectedEpoch = 504;
  //     const expectedHeight = 516096;
  //     const expectedSafeToProceedConsensus = true;
  //     await mineBlocks(
  //       (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
  //     );
  //     await expect(
  //       fixture.snapshots
  //         .connect(await getValidatorEthAccount(validatorsSnapshots[0]))
  //         .snapshot(signedData[503].GroupSignature, signedData[503].BClaims)
  //     )
  //       .to.emit(fixture.snapshots, `SnapshotTaken`)
  //       .withArgs(
  //         expectedChainId,
  //         expectedEpoch,
  //         expectedHeight,
  //         ethers.utils.getAddress(validatorsSnapshots[0].address),
  //         expectedSafeToProceedConsensus,
  //         signedData[503].GroupSignature
  //       );
  //   });
});
