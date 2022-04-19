import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import {
  assertEventValidatorMemberAdded,
  assertEventValidatorSetCompleted,
} from "../ethdkg/setup";
import { expect, Fixture, getFixture, mineBlocks } from "../setup";

describe("State Migration: Migrate state", () => {
  let fixture: Fixture;

  beforeEach(async function () {
    fixture = await getFixture();
  });

  it("Should migrate using special migration contract", async function () {
    const stateMigration = await (
      await ethers.getContractFactory("StateMigration")
    ).deploy(fixture.factory.address);

    const receipt1 = await (
      await fixture.factory.delegateCallAny(
        stateMigration.address,
        stateMigration.interface.encodeFunctionData("doMigrationStep")
      )
    ).wait();

    expect(receipt1.status).to.be.equals(1);

    await mineBlocks(2n);

    const txResponsePromise = fixture.factory.delegateCallAny(
      stateMigration.address,
      stateMigration.interface.encodeFunctionData("doMigrationStep")
    );
    const txResponse = await txResponsePromise;
    const receipt = await txResponse.wait();
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
        BigNumber.from(
          "0x109f68dde37442959baa4b16498a6fd19c285f9355c23d8eef900876e8536a12"
        ),
        BigNumber.from(
          "0x2c11cec2ce4e17afffcc105f9bd0e646f6274f562f1c93f93545fc22c74a2cdc"
        ),
        BigNumber.from(
          "0x0430024aa1619a117e74425481c44f4628f45af7b389e4d5f84fc41227e1829e"
        ),
        BigNumber.from(
          "0x163cb9abb41800ba5cc1955fd72c0983edc9869a21925006e691d4947451c9fd"
        ),
      ],
      [
        BigNumber.from(
          "0x13f8a33ff7ef3cb5536b2223195b7b652a3533d309ad3887bcf570c9b1dbe142"
        ),
        BigNumber.from(
          "0x170fe500681ff96e84a6dd7d1e4698f0ad6dd3ef17520a7cac29b29b84f86aa7"
        ),
        BigNumber.from(
          "0x19d29ec38a1d7d8d7284a76b214bf5b818eaccf47cd37ab8bc08c20833e586e9"
        ),
        BigNumber.from(
          "0x27c0f981d1bbc1667ea520341b7fa65fc79815ab8122a814e3714bfdbacd84db"
        ),
      ],
      [
        BigNumber.from(
          "0x1064dd800716a7e80ed5d40d5563940cd25be4451976a28c237e1426d40eae5a"
        ),
        BigNumber.from(
          "0x0e1c4d27ca7672e0662aaecbc5f2d62ec23e58cf63ffa9b6fabd41d3cff7c927"
        ),
        BigNumber.from(
          "0x106d1f91c4b77d5c9bb485aeea784e9acf0c91702eefb766e94aefc92043a004"
        ),
        BigNumber.from(
          "0x281971fd391a560142b8d796018afc31131a668b2ca6f62b304564d6422bb03f"
        ),
      ],
      [
        BigNumber.from(
          "0x2228b7dd85ddae13994fa85f42df1833da3b9468a1e65b987142d62f125a9754"
        ),
        BigNumber.from(
          "0x0c5682ae7cd22a3c3daff06ce469f318025845e90254d9d05cecaeba45f445a5"
        ),
        BigNumber.from(
          "0x2cdac99ed82ffc83fc17213e96d56400db23f08d05418936cb352f0e179cf971"
        ),
        BigNumber.from(
          "0x06371376125bb2b96a5e427fac829f5c3919296aac4c42ddc44eb7c773369c2b"
        ),
      ],
    ];

    const expectedValidatorCount = 4;
    const expectedNonce = 1;
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

    let ethLog = 0;
    for (let i = 0; i < receipt.logs.length; i++) {
      if (
        receipt.logs[i].topics[0].toLowerCase() ===
        "0x09b90b08bbc3dbe22e9d2a0bc9c2c7614c7511cd0ad72177727a1e762115bf06"
      ) {
        ethLog = i;
        break;
      }
    }
    for (let i = 0; i < validatorsAccounts.length; i++) {
      await assertEventValidatorMemberAdded(
        txResponse,
        validatorsAccounts[i],
        validatorIndexes[i],
        expectedNonce,
        expectedEpoch,
        [
          validatorShares[i][0],
          validatorShares[i][1],
          validatorShares[i][2],
          validatorShares[i][3],
        ],
        ethLog + i
      );
      const participantData = await fixture.ethdkg.getParticipantInternalState(
        validatorsAccounts[i]
      );
      expect(participantData.gpkj).to.be.deep.equals(
        validatorShares[i],
        "Incorrect gpkj"
      );
      expect(participantData.nonce).to.be.equal(
        expectedNonce,
        "Incorrect nonce"
      );
      expect(participantData.index).to.be.equal(
        validatorIndexes[i],
        "Incorrect index"
      );
      expect(
        await fixture.validatorPool.isValidator(validatorsAccounts[i])
      ).to.be.equals(true);
    }

    await assertEventValidatorSetCompleted(
      txResponse,
      expectedValidatorCount,
      expectedNonce,
      expectedEpoch,
      expectedEthHeight,
      expectedSideChainHeight,
      [
        masterPublicKey[0],
        masterPublicKey[1],
        masterPublicKey[2],
        masterPublicKey[3],
      ],
      ethLog + 4
    );

    expect(receipt.status).to.be.equals(1, "receipt failed");
    expect(await fixture.snapshots.getEpoch()).to.be.equal(79);
    expect(await fixture.ethdkg.getMasterPublicKeyHash()).to.be.equals(
      ethers.utils.solidityKeccak256(["uint256[4]"], [masterPublicKey]),
      "MPKs dont match!"
    );
    expect(await fixture.ethdkg.getMasterPublicKey()).to.be.deep.equals(
      [
        BigNumber.from(masterPublicKey[0]),
        BigNumber.from(masterPublicKey[1]),
        BigNumber.from(masterPublicKey[2]),
        BigNumber.from(masterPublicKey[3]),
      ],
      "MPKs2 dont match!"
    );
    expect(await fixture.ethdkg.getNumParticipants()).to.be.equals(
      expectedValidatorCount,
      "Incorrect num participant!"
    );
    expect(await fixture.ethdkg.getNonce()).to.be.equals(1, "Incorrect nonce!");
  });
});
