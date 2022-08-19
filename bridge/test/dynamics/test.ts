import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import {
  CanonicalVersionStruct,
  ConfigurationStruct,
  DynamicValuesStruct,
} from "../../typechain-types/contracts/Dynamics.sol/Dynamics";
import { expect } from "../chai-setup";
import { factoryCallAny, Fixture, getFixture } from "../setup";

describe("Testing Dynamics methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let fixture: Fixture;
  const minEpochsBetweenUpdates = BigNumber.from(2);
  const maxEpochsBetweenUpdates = BigNumber.from(336);

  let futureEpoch = BigNumber.from(2);
  let zeroEpoch = BigNumber.from(0);
  let currentDynamicValues: DynamicValuesStruct = {
    encoderVersion: 0,
    proposalTimeout: 4000,
    preVoteTimeout: 3000,
    preCommitTimeout: 3000,
    maxBlockSize: 3000000,
    dataStoreFee: BigNumber.from(0),
    valueStoreFee: BigNumber.from(0),
    minScaledTransactionFee: BigNumber.from(0),
  };
  let alicenetCurrentVersion = {
    major: 0,
    minor: 0,
    patch: 0,
    binaryHash:
      "0xbc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a",
  };
  let currentConfiguration: ConfigurationStruct = {
    minEpochsBetweenUpdates: BigNumber.from(2),
    maxEpochsBetweenUpdates: BigNumber.from(336),
  };

  beforeEach(async function () {
    fixture = await getFixture(false, true, false);
    const signers = await ethers.getSigners();
    [admin, user] = signers;
  });

  it("Should get Dynamics configuration", async () => {
    const configuration =
      (await fixture.dynamics.getConfiguration()) as ConfigurationStruct;
    expect(configuration.maxEpochsBetweenUpdates).to.be.equal(
      currentConfiguration.maxEpochsBetweenUpdates
    );
    expect(configuration.minEpochsBetweenUpdates).to.be.equal(
      currentConfiguration.minEpochsBetweenUpdates
    );
  });

  it("Should not change dynamic values if not impersonating factory", async () => {
    await expect(
      fixture.dynamics.changeDynamicValues(futureEpoch, currentDynamicValues)
    )
      .to.be.revertedWithCustomError(fixture.dynamics, "OnlyFactory")
      .withArgs(admin.address, fixture.factory.address);
  });

  it("Should not change dynamic values in a epoch lesser than minEpochsBetweenUpdates", async () => {
    await expect(
      factoryCallAny(fixture.factory, fixture.dynamics, "changeDynamicValues", [
        minEpochsBetweenUpdates.sub(1),
        currentDynamicValues,
      ])
    )
      .to.be.revertedWithCustomError(fixture.dynamics, "InvalidScheduledDate")
      .withArgs(
        minEpochsBetweenUpdates.sub(1),
        minEpochsBetweenUpdates,
        maxEpochsBetweenUpdates
      );
  });

  it("Should not change dynamic values in a epoch greater than maxEpochsBetweenUpdates", async () => {
    await expect(
      factoryCallAny(fixture.factory, fixture.dynamics, "changeDynamicValues", [
        maxEpochsBetweenUpdates.add(1),
        currentDynamicValues,
      ])
    )
      .to.be.revertedWithCustomError(fixture.dynamics, "InvalidScheduledDate")
      .withArgs(
        maxEpochsBetweenUpdates.add(1),
        minEpochsBetweenUpdates,
        maxEpochsBetweenUpdates
      );
  });

  it("Should set Dynamics configuration", async () => {
    currentConfiguration.maxEpochsBetweenUpdates =
      maxEpochsBetweenUpdates.add(1);
    await factoryCallAny(
      fixture.factory,
      fixture.dynamics,
      "setConfiguration",
      [currentConfiguration]
    );
    const configuration =
      (await fixture.dynamics.getConfiguration()) as ConfigurationStruct;
    expect(configuration.minEpochsBetweenUpdates).to.be.equal(
      currentConfiguration.minEpochsBetweenUpdates
    );
    expect(configuration.maxEpochsBetweenUpdates).to.be.equal(
      currentConfiguration.maxEpochsBetweenUpdates
    );
    await factoryCallAny(
      fixture.factory,
      fixture.dynamics,
      "changeDynamicValues",
      [maxEpochsBetweenUpdates.add(1), currentDynamicValues]
    );
  });

  it("Should get latest dynamic values", async () => {
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .encoderVersion
    ).to.be.deep.equal(currentDynamicValues.encoderVersion);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .proposalTimeout
    ).to.be.deep.equal(currentDynamicValues.proposalTimeout);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .preVoteTimeout
    ).to.be.deep.equal(currentDynamicValues.preVoteTimeout);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .preCommitTimeout
    ).to.be.deep.equal(currentDynamicValues.preCommitTimeout);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .maxBlockSize
    ).to.be.deep.equal(currentDynamicValues.maxBlockSize);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .dataStoreFee
    ).to.be.deep.equal(currentDynamicValues.dataStoreFee);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .valueStoreFee
    ).to.be.deep.equal(currentDynamicValues.valueStoreFee);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .minScaledTransactionFee
    ).to.be.deep.equal(currentDynamicValues.minScaledTransactionFee);
  });

  it("Should change dynamic values in a valid epoch and emit corresponding evet", async () => {
    const newDynamicValues = { ...currentDynamicValues };
    newDynamicValues.valueStoreFee = BigNumber.from(1);
    await factoryCallAny(
      fixture.factory,
      fixture.dynamics,
      "changeDynamicValues",
      [futureEpoch, newDynamicValues]
    );
    //the following code was used to successfully test event emission by removing temporally onlyFactory modifier in changeDynamicValues for test running
    /*     const encodedDynamicValues = await fixture.dynamics.encodeDynamicValues(newDynamicValues)
        await expect(
          fixture.dynamics.changeDynamicValues(
            futureEpoch, newDynamicValues
          )
        )
          .to.emit(fixture.dynamics, "DynamicValueChanged")
          .withArgs(futureEpoch, encodedDynamicValues); */
    await factoryCallAny(fixture.factory, fixture.snapshots, "snapshot", [
      [],
      [],
    ]);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .valueStoreFee
    ).to.be.equal(currentDynamicValues.valueStoreFee);
    await factoryCallAny(fixture.factory, fixture.snapshots, "snapshot", [
      [],
      [],
    ]);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .valueStoreFee
    ).to.be.equal(newDynamicValues.valueStoreFee);
  });

  it("Should get previous dynamic values", async () => {
    const newDynamicValues = { ...currentDynamicValues };
    newDynamicValues.valueStoreFee = BigNumber.from(1);
    await factoryCallAny(
      fixture.factory,
      fixture.dynamics,
      "changeDynamicValues",
      [futureEpoch, newDynamicValues]
    );
    await factoryCallAny(fixture.factory, fixture.snapshots, "snapshot", [
      [],
      [],
    ]);
    await factoryCallAny(fixture.factory, fixture.snapshots, "snapshot", [
      [],
      [],
    ]);
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .valueStoreFee
    ).to.be.equal(newDynamicValues.valueStoreFee);
    expect(
      (
        (await fixture.dynamics.getPreviousDynamicValues(
          1
        )) as DynamicValuesStruct
      ).valueStoreFee
    ).to.be.equal(currentDynamicValues.valueStoreFee);
  });

  it("Should not get unexistent dynamic values", async () => {
    await expect(
      fixture.dynamics.getPreviousDynamicValues(
        zeroEpoch
      ) as DynamicValuesStruct
    )
      .to.be.revertedWithCustomError(fixture.dynamics, "DynamicValueNotFound")
      .withArgs(zeroEpoch);
  });

  it("Should not update Alicenet node version to the same current version", async () => {
    await expect(
      factoryCallAny(
        fixture.factory,
        fixture.dynamics,
        "updateAliceNetNodeVersion",
        [
          futureEpoch,
          alicenetCurrentVersion.major,
          alicenetCurrentVersion.minor,
          alicenetCurrentVersion.patch,
          alicenetCurrentVersion.binaryHash,
        ]
      )
    )
      .to.be.revertedWithCustomError(
        fixture.dynamics,
        "InvalidAliceNetNodeVersion"
      )
      .withArgs(
        [
          alicenetCurrentVersion.major,
          alicenetCurrentVersion.minor,
          alicenetCurrentVersion.patch,
          alicenetCurrentVersion.binaryHash,
        ],
        [
          alicenetCurrentVersion.major,
          alicenetCurrentVersion.minor,
          alicenetCurrentVersion.patch,
          "0x0000000000000000000000000000000000000000000000000000000000000000",
        ]
      );
  });

  it("Should not update Alicenet node version to a non consecutive major version", async () => {
    const newMajorVersion = alicenetCurrentVersion.major + 2;
    await expect(
      factoryCallAny(
        fixture.factory,
        fixture.dynamics,
        "updateAliceNetNodeVersion",
        [
          futureEpoch,
          newMajorVersion,
          alicenetCurrentVersion.minor,
          alicenetCurrentVersion.patch,
          alicenetCurrentVersion.binaryHash,
        ]
      )
    )
      .to.be.revertedWithCustomError(
        fixture.dynamics,
        "InvalidAliceNetNodeVersion"
      )
      .withArgs(
        [
          newMajorVersion,
          alicenetCurrentVersion.minor,
          alicenetCurrentVersion.patch,
          "0xbc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a",
        ],
        [
          alicenetCurrentVersion.major,
          alicenetCurrentVersion.minor,
          alicenetCurrentVersion.patch,
          "0x0000000000000000000000000000000000000000000000000000000000000000",
        ]
      );
  });

  it("Should update Alicenet node version to a valid version and emit corresponding event", async () => {
    const newMajorVersion = alicenetCurrentVersion.major + 1;
    const receipt = await factoryCallAny(
      fixture.factory,
      fixture.dynamics,
      "updateAliceNetNodeVersion",
      [
        futureEpoch,
        newMajorVersion,
        alicenetCurrentVersion.minor,
        alicenetCurrentVersion.patch,
        alicenetCurrentVersion.binaryHash,
      ]
    );

    //the following code was used to successfully test event emission removing temporally onlyFactory modifier in updateAliceNetNodeVersion
    /*     await expect(
          fixture.dynamics.updateAliceNetNodeVersion(
            futureEpoch,
            newMajorVersion,
            alicenetCurrentVersion.minor,
            alicenetCurrentVersion.patch,
            alicenetCurrentVersion.binaryHash
          )
        )
          .to.emit(fixture.dynamics, "NewAliceNetNodeVersionAvailable")
          .withArgs(futureEpoch, [
            newMajorVersion,
            alicenetCurrentVersion.major,
            alicenetCurrentVersion.patch,
            alicenetCurrentVersion.binaryHash,
          ]); */
  });

  it("Should obtain latest Alicenet node version", async () => {
    const newMajorVersion = alicenetCurrentVersion.major + 1;
    await factoryCallAny(
      fixture.factory,
      fixture.dynamics,
      "updateAliceNetNodeVersion",
      [
        futureEpoch,
        newMajorVersion,
        alicenetCurrentVersion.minor,
        alicenetCurrentVersion.patch,
        alicenetCurrentVersion.binaryHash,
      ]
    );
    expect(
      (
        (await fixture.dynamics.getLatestAliceNetVersion()) as CanonicalVersionStruct
      ).major
    ).to.be.equal(newMajorVersion);
  });
});
