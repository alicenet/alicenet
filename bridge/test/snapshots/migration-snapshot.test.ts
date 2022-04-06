import { BigNumber } from "ethers";
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
} from "../setup";
import { createValidators, stakeValidators } from "../validatorPool/setup";
import {
  invalidSnapshot500,
  validSnapshot1024,
} from "./assets/4-validators-snapshots-1";

describe("Snapshots: Migrate state", () => {
  let fixture: Fixture;
  let expectedBlockNumber: bigint;

  beforeEach(async function () {
    fixture = await getFixture();
  });

  it("Should not be to do a migration of snapshots if not factory", async function () {
    await expect(
      fixture.snapshots.migrateSnapshots(
        [validSnapshot1024.GroupSignature],
        [validSnapshot1024.BClaims]
      )
    ).to.be.revertedWith("2000");
  });

  it("Should not be to do a migration with mismatch data length", async function () {
    await expect(
      factoryCallAny(fixture.factory, fixture.snapshots, "migrateSnapshots", [
        [validSnapshot1024.GroupSignature],
        [validSnapshot1024.BClaims, validSnapshot1024.BClaims],
      ])
    ).to.be.revertedWith("409");
  });

  it("Should not be to do a migration with incorrect block height", async function () {
    await expect(
      factoryCallAny(fixture.factory, fixture.snapshots, "migrateSnapshots", [
        [invalidSnapshot500.GroupSignature, validSnapshot1024.GroupSignature],
        [invalidSnapshot500.BClaims, validSnapshot1024.BClaims],
      ])
    ).to.be.revertedWith("406");
    await expect(
      factoryCallAny(fixture.factory, fixture.snapshots, "migrateSnapshots", [
        [validSnapshot1024.GroupSignature, invalidSnapshot500.GroupSignature],
        [validSnapshot1024.BClaims, invalidSnapshot500.BClaims],
      ])
    ).to.be.revertedWith("406");
    await expect(
      factoryCallAny(fixture.factory, fixture.snapshots, "migrateSnapshots", [
        [invalidSnapshot500.GroupSignature],
        [invalidSnapshot500.BClaims],
      ])
    ).to.be.revertedWith("406");
  });

  it("After a migration we should not be able to change validators and rerun ethdkg without scheduling maintenance", async function () {
    let expectedChainId = 1;
    let expectedEpoch = 1;
    let expectedHeight = validSnapshot1024.height as number;
    let expectedSafeToProceedConsensus = false;
    const fixture = await getFixture();
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
      fixture.snapshots,
      "migrateSnapshots",
      [
        [signedData[500].GroupSignature, signedData[501].GroupSignature],
        [signedData[500].BClaims, signedData[501].BClaims],
      ]
    );

    expect(receipt.status).to.be.equals(1);

    // validatorsAccounts_,
    //     uint256[] memory validatorIndexes_,
    //     uint256[4][] memory validatorShares_,
    //     uint8 validatorCount_,
    //     uint256 epoch_,
    //     uint256 sideChainHeight_,
    //     uint256 ethHeight_,
    //     uint256[4] memory masterPublicKey_
    receipt = await factoryCallAny(
      fixture.factory,
      fixture.ethdkg,
      "migrateValidators",
      [,]
    );

    // await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    // await completeETHDKGRound(validatorsSnapshots, {
    //   ethdkg: fixture.ethdkg,
    //   validatorPool: fixture.validatorPool,
    // });
    // await factoryCallAnyFixture(
    //   fixture,
    //   "validatorPool",
    //   "scheduleMaintenance"
    // );
    // await mineBlocks(
    //   (await snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    // );
    // await expect(
    //   snapshots
    //     .connect(await getValidatorEthAccount(validatorsSnapshots1[0]))
    //     .snapshot(validSnapshot1024.GroupSignature, validSnapshot1024.BClaims)
    // )
    //   .to.emit(snapshots, `SnapshotTaken`)
    //   .withArgs(
    //     expectedChainId,
    //     expectedEpoch,
    //     expectedHeight,
    //     ethers.utils.getAddress(validatorsSnapshots1[0].address),
    //     expectedSafeToProceedConsensus,
    //     validSnapshot1024.GroupSignature
    //   );
    // await factoryCallAnyFixture(
    //   fixture,
    //   "validatorPool",
    //   "unregisterValidators",
    //   [validators]
    // );

    // // registering the new validators
    // const newValidators = await createValidators(fixture, validatorsSnapshots2);
    // const newStakingTokenIds = await stakeValidators(fixture, newValidators);
    // await factoryCallAnyFixture(
    //   fixture,
    //   "validatorPool",
    //   "registerValidators",
    //   [newValidators, newStakingTokenIds]
    // );
    // await factoryCallAnyFixture(fixture, "validatorPool", "initializeETHDKG");
    // await completeETHDKGRound(
    //   validatorsSnapshots2,
    //   {
    //     ethdkg: fixture.ethdkg,
    //     validatorPool: fixture.validatorPool,
    //   },
    //   expectedEpoch,
    //   expectedHeight,
    //   (await snapshots.getCommittedHeightFromLatestSnapshot()).toNumber()
    // );
    // await mineBlocks(
    //   (await fixture.snapshots.getMinimumIntervalBetweenSnapshots()).toBigInt()
    // );
    // expectedChainId = 1;
    // expectedEpoch = 2;
    // expectedHeight = validSnapshot2048.height as number;
    // expectedSafeToProceedConsensus = true;
    // await expect(
    //   snapshots
    //     .connect(await getValidatorEthAccount(validatorsSnapshots2[0]))
    //     .snapshot(validSnapshot2048.GroupSignature, validSnapshot2048.BClaims)
    // )
    //   .to.emit(snapshots, `SnapshotTaken`)
    //   .withArgs(
    //     expectedChainId,
    //     expectedEpoch,
    //     expectedHeight,
    //     ethers.utils.getAddress(validatorsSnapshots2[0].address),
    //     expectedSafeToProceedConsensus,
    //     validSnapshot2048.GroupSignature
    //   );
  });

  it("Should migrate using special migration contract", async function () {
    const stateMigration = await (
      await ethers.getContractFactory("StateMigration")
    ).deploy(fixture.factory.address);

    const txResponsePromise = fixture.factory.delegateCallAny(
      stateMigration.address,
      stateMigration.interface.encodeFunctionData(
        "migrateSnapshotsAndValidators"
      )
    );
    const expectedChainId = 1;
    const expectedAliceNetHeights = [76800n, 77824n, 78848n, 79872n, 80896n];
    const expectedEpochs = [75n, 76n, 77n, 78n, 79n];
    const signatures = [
      "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f114551b8239e68c2fc16a68bbfbfe2140b718ca279d784074743ce1dcdb134ed10d0a4d630460957d1c50c0e3a8238cafc3985651674ce03e4b91837da6080de6",
      "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f11a4d9a0e85b1e265f221c163546d61fcf76b301944368abbfbba42dc56a083ba2ac800dc9a20a25ca95146c65d6c6cddbb299625907c1a057754f70073ec8675",
      "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f106e2ca23e60db68ff939899c926fd9d76e40d15b17720bd5d60df4fd9725cd07288ca12870d4b48f441e6a5b1943c8b9c91f0bd28256ab352e77a61d23124dbb",
      "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f11bbb68f54eb8ab7b8276432c152909f11ba49cf685c07fadc1e1ba96c1b579ee27002d8fe6bf013b640e1904525645c5f481cc47358330a8b6eb29d019828e33",
      "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f1046a12b7354767f6ec2e660540eee970333bfa01e458ee4cd066588d3c4632972e3c60a8f58a5f89b0926ae265b921bed31fc830056980d70e58db642357af02",
    ];

    for (let i = 0; i < signatures.length; i++) {
      await expect(txResponsePromise, "Failed to found event for " + i)
        .to.emit(fixture.snapshots, `SnapshotTaken`)
        .withArgs(
          expectedChainId,
          expectedEpochs[i],
          expectedAliceNetHeights[i],
          fixture.factory.address,
          true,
          signatures[i]
        );
    }

    const validatorsAccounts = [
      ethers.utils.getAddress("0xb80d6653f7e5b80dbbe8d0aa9f61b5d72e8028ad"),
      ethers.utils.getAddress("0x25489d6a720663f7e5253df68948edb302dfdcb6"),
      ethers.utils.getAddress("0x322e8f463b925da54a778ed597aef41bc4fe4743"),
      ethers.utils.getAddress("0xadf2a338e19c12298a3007cbea1c5276d1f746e0"),
    ];

    const validatorIndexes = [
      BigNumber.from(
        "0x0000000000000000000000000000000000000000000000000000000000000001"
      ),
      BigNumber.from(
        "0x0000000000000000000000000000000000000000000000000000000000000002"
      ),
      BigNumber.from(
        "0x0000000000000000000000000000000000000000000000000000000000000003"
      ),
      BigNumber.from(
        "0x0000000000000000000000000000000000000000000000000000000000000004"
      ),
    ];
    const validatorShares = [
      [
        "0x109f68dde37442959baa4b16498a6fd19c285f9355c23d8eef900876e8536a12",
        "0x2c11cec2ce4e17afffcc105f9bd0e646f6274f562f1c93f93545fc22c74a2cdc",
        "0x0430024aa1619a117e74425481c44f4628f45af7b389e4d5f84fc41227e1829e",
        "0x163cb9abb41800ba5cc1955fd72c0983edc9869a21925006e691d4947451c9fd",
      ],
      [
        "0x13f8a33ff7ef3cb5536b2223195b7b652a3533d309ad3887bcf570c9b1dbe142",
        "0x170fe500681ff96e84a6dd7d1e4698f0ad6dd3ef17520a7cac29b29b84f86aa7",
        "0x19d29ec38a1d7d8d7284a76b214bf5b818eaccf47cd37ab8bc08c20833e586e9",
        "0x27c0f981d1bbc1667ea520341b7fa65fc79815ab8122a814e3714bfdbacd84db",
      ],
      [
        "0x1064dd800716a7e80ed5d40d5563940cd25be4451976a28c237e1426d40eae5a",
        "0x0e1c4d27ca7672e0662aaecbc5f2d62ec23e58cf63ffa9b6fabd41d3cff7c927",
        "0x106d1f91c4b77d5c9bb485aeea784e9acf0c91702eefb766e94aefc92043a004",
        "0x281971fd391a560142b8d796018afc31131a668b2ca6f62b304564d6422bb03f",
      ],
      [
        "0x2228b7dd85ddae13994fa85f42df1833da3b9468a1e65b987142d62f125a9754",
        "0x0c5682ae7cd22a3c3daff06ce469f318025845e90254d9d05cecaeba45f445a5",
        "0x2cdac99ed82ffc83fc17213e96d56400db23f08d05418936cb352f0e179cf971",
        "0x06371376125bb2b96a5e427fac829f5c3919296aac4c42ddc44eb7c773369c2b",
      ],
    ];

    const txResponse = await txResponsePromise;
    const receipt = await txResponse.wait();

    const expectedValidatorCount = 4;
    const expectedNonce = 0;
    const expectedEpoch = 1;
    const expectedEthHeight = 0x236;
    const expectedSideChainHeight = 0;
    const masterPublicKey = [
      BigInt(
        "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877"
      ),
      BigInt(
        "0x081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3"
      ),
      BigInt(
        "0x253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3"
      ),
      BigInt(
        "0x095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f1"
      ),
    ];

    // for (let i = 0; i < validatorsAccounts.length; i++) {
    //   //   await expect(
    //   //     txResponsePromise,
    //   //     "Failed to Validator member found event for " + i
    //   //   )
    //   //     .to.emit(fixture.ethdkg, "ValidatorMemberAdded")
    //   //     .withArgs(
    //   //       validatorsAccounts[i],
    //   //       validatorIndexes[i],
    //   //       expectedEpoch,
    //   //       validatorShares[i][0],
    //   //       validatorShares[i][1],
    //   //       validatorShares[i][2],
    //   //       validatorShares[i][3]
    //   //     );
    //   const myEvent = fixture.ethdkg.interface.decodeEventLog(
    //     "ValidatorMemberAdded",
    //     receipt.logs[i + 5].data,
    //     receipt.logs[i + 5].topics
    //   );
    //   await expect(myEvent).to.be.deep.equals({
    //     account: validatorsAccounts[i],
    //     index: validatorIndexes[i],
    //     expectedEpoch,
    //   });
    // }

    await expect(
      txResponsePromise,
      "Failed to ValidatorSetCompleted found event for "
    )
      .to.emit(fixture.snapshots, "ValidatorSetCompleted")
      .withArgs(
        expectedValidatorCount,
        expectedNonce,
        expectedEpoch,
        expectedEthHeight,
        expectedSideChainHeight,
        masterPublicKey[0],
        masterPublicKey[1],
        masterPublicKey[2],
        masterPublicKey[3]
      );

    expect(receipt.status).to.be.equals(1);
    expect(await fixture.snapshots.getEpoch()).to.be.equal(79);
    expect(await fixture.ethdkg.getMasterPublicKeyHash()).to.be.equals(
      ethers.utils.solidityKeccak256(["uint256[4]"], [masterPublicKey]),
      "MPKs dont match!"
    );
  });

  describe("After a successful migration", () => {
    beforeEach(async function () {
      const bClaims = [
        "0x000000000100040015000000002c01000d00000002010000190000000201000025000000020100003100000002010000031dfcf2fef268ff9956ee399230e9bf1da9dd510d18552b736b3269f4544c01c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470d058c043d927976ecf061b3cdb3a4a0d2de3284fcd69e23733650a1b3ef367533807ec1e085227e7bb99f47db1b118cefdae66f2fbfc66449a4500e9a6a2dab2",
        "0x000000000100040015000000003001000d000000020100001900000002010000250000000201000031000000020100009a7a9e6d46b1640392f4444a9cf56d1190fe77fd4a740ee76b0e5f261341d195c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470d29f86626d42e94c88da05e5cec3c29f0fd037f8a9e1fcb6b49a4dd322da317ce4c870a97b5731173a6d17b71740c498ed409e25e28e9077c7f9119af3c28692",
        "0x000000000100040015000000003401000d0000000201000019000000020100002500000002010000310000000201000000f396eeda71abea614606937f7fcbd4d704af9ac0556a66687d689497c8da09c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47033839738f138dbcbb362c3b351c7b7f16041304c75354fb11ae01d3623cc4e146a5a9af572eacd9e40d9f508d077419cc191f542c213d2c204d3251ce88c476b",
        "0x000000000100040015000000003801000d0000000201000019000000020100002500000002010000310000000201000000af33d9a061b001d8c1c912b2cf58f5f5bccd81e9c0fac7ac4f256134677a27c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a47099726b1e813baf97a0f88c89c5257358c4ef40c38b515184ea95bb9113587c85a06879b5886d1af4f04773c418b9517db8b410de7fdff0fd9ed47316e4c23e9f",
        "0x000000000100040015000000003c01000d000000020100001900000002010000250000000201000031000000020100001923548c43980ec331fa993cb8b90b157f4251fc8c37ba3506d205611af468e8c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a4702df6fa1cfdeabd709149817a42eb2c1e2c18cc06c6b2bbf4a51d825aaa442f3516f8b5f4a60397c0efdd38750282135beff68f4cdff36497574894658e2807ce",
      ];

      const signatures = [
        "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f114551b8239e68c2fc16a68bbfbfe2140b718ca279d784074743ce1dcdb134ed10d0a4d630460957d1c50c0e3a8238cafc3985651674ce03e4b91837da6080de6",
        "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f11a4d9a0e85b1e265f221c163546d61fcf76b301944368abbfbba42dc56a083ba2ac800dc9a20a25ca95146c65d6c6cddbb299625907c1a057754f70073ec8675",
        "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f106e2ca23e60db68ff939899c926fd9d76e40d15b17720bd5d60df4fd9725cd07288ca12870d4b48f441e6a5b1943c8b9c91f0bd28256ab352e77a61d23124dbb",
        "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f11bbb68f54eb8ab7b8276432c152909f11ba49cf685c07fadc1e1ba96c1b579ee27002d8fe6bf013b640e1904525645c5f481cc47358330a8b6eb29d019828e33",
        "0x2c28ce7f0c752e035b68687a8210cceb6068b5034bba9a4a8f2d43e3bbaa8877081b33b885370e04cd712601eb860bf821396bdbcd4b089aba0bfe7b1e649dd3253adba688741303e0b046632b35289a0d5c7648b414375e4d61a855abc5f0c3095ed894617e232df1779101e1d98e177340cb0fc6283cbc437d79a12290c2f1046a12b7354767f6ec2e660540eee970333bfa01e458ee4cd066588d3c4632972e3c60a8f58a5f89b0926ae265b921bed31fc830056980d70e58db642357af02",
      ];
      const receipt = await factoryCallAny(
        fixture.factory,
        fixture.snapshots,
        "migrateSnapshots",
        [signatures, bClaims]
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
          [validSnapshot1024.GroupSignature],
          [validSnapshot1024.BClaims],
        ])
      ).to.be.revertedWith("408");
    });
  });
});
