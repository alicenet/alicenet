import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import {
  CanonicalVersionStruct,
  CanonicalVersionStructOutput,
  ConfigurationStruct,
  DynamicValuesStruct,
} from "../../typechain-types/contracts/Dynamics";
import { expect } from "../chai-setup";
import { factoryCallAny, Fixture, getFixture } from "../setup";
import { commitSnapshots } from "../validatorPool/setup";

const changeDynamicValues = async (
  fixture: Fixture,
  newDynamicValues: DynamicValuesStruct
) => {
  const minEpochsBetweenUpdates = 2;
  const addToCurrentSnapshotEpoch = 10;
  const encodedDynamicValues = await fixture.dynamics.encodeDynamicValues(
    newDynamicValues
  );
  // mine some blocks to have time progression
  await commitSnapshots(fixture, addToCurrentSnapshotEpoch);
  await expect(
    fixture.factory.callAny(
      fixture.dynamics.address,
      0,
      fixture.dynamics.interface.encodeFunctionData("changeDynamicValues", [
        minEpochsBetweenUpdates,
        newDynamicValues,
      ])
    )
  )
    .to.emit(fixture.dynamics, "DynamicValueChanged")
    .withArgs(
      (await fixture.snapshots.getEpoch()).add(minEpochsBetweenUpdates),
      encodedDynamicValues
    );
};

const updateAliceNetNode = async (
  fixture: Fixture,
  newAliceNetVersion: CanonicalVersionStructOutput,
  relativeEpoch: number,
  assertReturn: boolean,
  assertSnapshotEvent: boolean
): Promise<any> => {
  const addToCurrentSnapshotEpoch = 10;
  await commitSnapshots(fixture, addToCurrentSnapshotEpoch);
  const txPromise = fixture.factory.callAny(
    fixture.dynamics.address,
    0,
    fixture.dynamics.interface.encodeFunctionData("updateAliceNetNodeVersion", [
      relativeEpoch,
      newAliceNetVersion.major,
      newAliceNetVersion.minor,
      newAliceNetVersion.patch,
      newAliceNetVersion.binaryHash,
    ])
  );
  const expectedEpoch = (await fixture.snapshots.getEpoch()).add(relativeEpoch);
  if (assertReturn) {
    await expect(txPromise)
      .to.emit(fixture.dynamics, "NewAliceNetNodeVersionAvailable")
      .withArgs([
        newAliceNetVersion.major,
        newAliceNetVersion.minor,
        newAliceNetVersion.patch,
        expectedEpoch,
        newAliceNetVersion.binaryHash,
      ]);
  }

  if (assertSnapshotEvent) {
    await fixture.snapshots.snapshot("0x00", "0x00");
    await expect(fixture.snapshots.snapshot("0x00", "0x00"))
      .to.emit(fixture.dynamics, "NewCanonicalAliceNetNodeVersion")
      .withArgs([
        newAliceNetVersion.major,
        newAliceNetVersion.minor,
        newAliceNetVersion.patch,
        expectedEpoch,
        newAliceNetVersion.binaryHash,
      ]);
  }
  return txPromise;
};

const shouldFailUpdateAliceNetNode = async (
  fixture: Fixture,
  newAliceNetVersion: CanonicalVersionStructOutput
) => {
  const minEpochsBetweenUpdates = 2;
  const alicenetCurrentVersion =
    await fixture.dynamics.getLatestAliceNetVersion();
  const expectedEpoch = (await fixture.snapshots.getEpoch()).add(
    minEpochsBetweenUpdates
  );
  await expect(
    fixture.factory.callAny(
      fixture.dynamics.address,
      0,
      fixture.dynamics.interface.encodeFunctionData(
        "updateAliceNetNodeVersion",
        [
          minEpochsBetweenUpdates,
          newAliceNetVersion.major,
          newAliceNetVersion.minor,
          newAliceNetVersion.patch,
          newAliceNetVersion.binaryHash,
        ]
      )
    )
  )
    .to.be.revertedWithCustomError(
      fixture.dynamics,
      "InvalidAliceNetNodeVersion"
    )
    .withArgs(
      [
        newAliceNetVersion.major,
        newAliceNetVersion.minor,
        newAliceNetVersion.patch,
        expectedEpoch,
        newAliceNetVersion.binaryHash,
      ],
      [
        alicenetCurrentVersion.major,
        alicenetCurrentVersion.minor,
        alicenetCurrentVersion.patch,
        alicenetCurrentVersion.executionEpoch,
        alicenetCurrentVersion.binaryHash,
      ]
    );
};

describe("Testing Dynamics methods", async () => {
  let admin: SignerWithAddress;
  let fixture: Fixture;
  const minEpochsBetweenUpdates = BigNumber.from(2);
  const maxEpochsBetweenUpdates = BigNumber.from(336);
  const zeroEpoch = BigNumber.from(0);
  const currentDynamicValues: DynamicValuesStruct = {
    encoderVersion: 0,
    proposalTimeout: 4000,
    preVoteTimeout: 3000,
    preCommitTimeout: 3000,
    maxBlockSize: 3000000,
    dataStoreFee: BigNumber.from(0),
    valueStoreFee: BigNumber.from(0),
    minScaledTransactionFee: BigNumber.from(0),
  };
  const currentConfiguration: ConfigurationStruct = {
    minEpochsBetweenUpdates: BigNumber.from(2),
    maxEpochsBetweenUpdates: BigNumber.from(336),
  };
  let alicenetCurrentVersion: CanonicalVersionStructOutput;

  beforeEach(async function () {
    fixture = await getFixture(false, true, false);
    const signers = await ethers.getSigners();
    [admin] = signers;
    alicenetCurrentVersion = {
      ...(await fixture.dynamics.getLatestAliceNetVersion()),
    };
    alicenetCurrentVersion.binaryHash =
      "0xbc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a";
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
      fixture.dynamics.changeDynamicValues(
        minEpochsBetweenUpdates,
        currentDynamicValues
      )
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
    const latestDynamicsValue =
      (await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct;
    expect(latestDynamicsValue.encoderVersion).to.be.deep.equal(
      currentDynamicValues.encoderVersion
    );
    expect(latestDynamicsValue.proposalTimeout).to.be.deep.equal(
      currentDynamicValues.proposalTimeout
    );
    expect(latestDynamicsValue.preVoteTimeout).to.be.deep.equal(
      currentDynamicValues.preVoteTimeout
    );
    expect(latestDynamicsValue.preCommitTimeout).to.be.deep.equal(
      currentDynamicValues.preCommitTimeout
    );
    expect(latestDynamicsValue.maxBlockSize).to.be.deep.equal(
      currentDynamicValues.maxBlockSize
    );
    expect(latestDynamicsValue.dataStoreFee).to.be.deep.equal(
      currentDynamicValues.dataStoreFee
    );
    expect(latestDynamicsValue.valueStoreFee).to.be.deep.equal(
      currentDynamicValues.valueStoreFee
    );
    expect(latestDynamicsValue.minScaledTransactionFee).to.be.deep.equal(
      currentDynamicValues.minScaledTransactionFee
    );
  });

  it("Should change dynamic values in a valid epoch and emit corresponding event", async () => {
    const newDynamicValues = { ...currentDynamicValues };
    newDynamicValues.valueStoreFee = BigNumber.from(1);
    await changeDynamicValues(fixture, newDynamicValues);
    // before the epochs has passed the value should be the same
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .valueStoreFee
    ).to.be.equal(currentDynamicValues.valueStoreFee);
    await commitSnapshots(fixture, minEpochsBetweenUpdates.toNumber());
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .valueStoreFee
    ).to.be.equal(newDynamicValues.valueStoreFee);
  });

  it("Should get previous dynamic values", async () => {
    const newDynamicValues = { ...currentDynamicValues };
    newDynamicValues.valueStoreFee = BigNumber.from(1);
    await changeDynamicValues(fixture, newDynamicValues);
    await commitSnapshots(fixture, minEpochsBetweenUpdates.toNumber());
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

  it("Should not get inexistent previous dynamic values", async () => {
    await expect(fixture.dynamics.getPreviousDynamicValues(zeroEpoch))
      .to.be.revertedWithCustomError(fixture.dynamics, "DynamicValueNotFound")
      .withArgs(zeroEpoch);
  });

  it("Should not get dynamic values before the previous node", async () => {
    const newDynamicValues = { ...currentDynamicValues };
    newDynamicValues.valueStoreFee = BigNumber.from(1);
    await commitSnapshots(fixture, 100);
    await changeDynamicValues(fixture, newDynamicValues);
    await commitSnapshots(fixture, minEpochsBetweenUpdates.toNumber());
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .valueStoreFee
    ).to.be.equal(newDynamicValues.valueStoreFee);
    await commitSnapshots(fixture, 100);
    const newDynamicValues2 = { ...newDynamicValues };
    newDynamicValues2.maxBlockSize = BigNumber.from("400000000");
    await changeDynamicValues(fixture, newDynamicValues2);
    await commitSnapshots(fixture, minEpochsBetweenUpdates.toNumber());
    expect(
      ((await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct)
        .maxBlockSize
    ).to.be.equal(newDynamicValues2.maxBlockSize);

    // we are close to epoch 200, previous value changed close to epoch 100
    await expect(fixture.dynamics.getPreviousDynamicValues(30))
      .to.be.revertedWithCustomError(fixture.dynamics, "DynamicValueNotFound")
      .withArgs(30);
  });

  it("Should update AliceNet node version to a valid version and emit corresponding event", async () => {
    const newAliceNetVersion = {
      ...alicenetCurrentVersion,
    };
    newAliceNetVersion.major += 1;
    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      true
    );
  });

  it("Should update AliceNet node version to a valid version and emit corresponding event", async () => {
    const newAliceNetVersion = {
      ...alicenetCurrentVersion,
    };
    newAliceNetVersion.major += 1;
    newAliceNetVersion.binaryHash = ethers.utils.formatBytes32String("0xaa");
    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      true
    );

    newAliceNetVersion.minor += 1;
    newAliceNetVersion.binaryHash = ethers.utils.formatBytes32String("0xbb");
    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      true
    );

    newAliceNetVersion.patch += 1;
    newAliceNetVersion.binaryHash = ethers.utils.formatBytes32String("0xcc");
    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      true
    );

    newAliceNetVersion.minor += 1;
    newAliceNetVersion.patch -= 1;
    newAliceNetVersion.binaryHash = ethers.utils.formatBytes32String("0xdd");
    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      true
    );

    newAliceNetVersion.major += 1;
    newAliceNetVersion.minor -= 1;
    newAliceNetVersion.patch += 1;
    newAliceNetVersion.binaryHash = ethers.utils.formatBytes32String("0xde");
    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      true
    );
  });

  it("Should obtain latest AliceNet node version", async () => {
    const newAliceNetVersion = {
      ...alicenetCurrentVersion,
    };
    newAliceNetVersion.major += 1;

    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      false
    );
    const latestNode =
      (await fixture.dynamics.getLatestAliceNetVersion()) as CanonicalVersionStruct;
    expect(latestNode.major).to.be.equal(newAliceNetVersion.major);
    expect(latestNode.minor).to.be.equal(newAliceNetVersion.minor);
    expect(latestNode.patch).to.be.equal(newAliceNetVersion.patch);
    expect(latestNode.executionEpoch).to.be.equal(
      minEpochsBetweenUpdates.toNumber() +
        (await fixture.snapshots.getEpoch()).toNumber()
    );
    expect(latestNode.binaryHash).to.be.equal(newAliceNetVersion.binaryHash);
  });

  it("Should not update AliceNet node version to smaller version", async () => {
    const newAliceNetVersion = {
      ...alicenetCurrentVersion,
    };
    newAliceNetVersion.major += 1;
    newAliceNetVersion.minor += 1;
    newAliceNetVersion.patch += 1;

    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      true
    );
    let pastAliceNet = newAliceNetVersion;
    pastAliceNet.minor -= 1;
    await shouldFailUpdateAliceNetNode(fixture, pastAliceNet);

    pastAliceNet = newAliceNetVersion;
    pastAliceNet.patch -= 1;
    await shouldFailUpdateAliceNetNode(fixture, pastAliceNet);

    pastAliceNet = newAliceNetVersion;
    pastAliceNet.major -= 1;
    await shouldFailUpdateAliceNetNode(fixture, pastAliceNet);
  });

  it("Should not update AliceNet node version to the same current version", async () => {
    await shouldFailUpdateAliceNetNode(
      fixture,
      await fixture.dynamics.getLatestAliceNetVersion()
    );
  });

  it("Should not update AliceNet node version to a non consecutive major version", async () => {
    const newAliceNetVersion = {
      ...alicenetCurrentVersion,
    };
    newAliceNetVersion.major += 2;
    const addToCurrentEpoch = 10;
    await commitSnapshots(fixture, addToCurrentEpoch);
    await shouldFailUpdateAliceNetNode(fixture, newAliceNetVersion);
  });

  it("Should not update AliceNet node with same hash", async () => {
    const newAliceNetVersion = {
      ...alicenetCurrentVersion,
    };
    newAliceNetVersion.major += 1;
    newAliceNetVersion.minor += 1;
    newAliceNetVersion.patch += 1;

    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      true
    );
    const addToCurrentEpoch = 10;
    newAliceNetVersion.major += 1;
    await commitSnapshots(fixture, addToCurrentEpoch);
    await expect(
      updateAliceNetNode(
        fixture,
        newAliceNetVersion,
        minEpochsBetweenUpdates.toNumber(),
        false,
        false
      )
    )
      .to.be.revertedWithCustomError(
        fixture.dynamics,
        "InvalidAliceNetNodeHash"
      )
      .withArgs(
        "0xbc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a",
        "0xbc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"
      );
  });

  it("Should not update AliceNet node hash 0", async () => {
    const newAliceNetVersion = {
      ...alicenetCurrentVersion,
    };
    newAliceNetVersion.major += 1;
    newAliceNetVersion.minor += 1;
    newAliceNetVersion.patch += 1;

    await updateAliceNetNode(
      fixture,
      newAliceNetVersion,
      minEpochsBetweenUpdates.toNumber(),
      true,
      true
    );
    const addToCurrentEpoch = 10;
    newAliceNetVersion.major += 1;
    newAliceNetVersion.binaryHash = ethers.utils.formatBytes32String("");
    await commitSnapshots(fixture, addToCurrentEpoch);
    await expect(
      updateAliceNetNode(
        fixture,
        newAliceNetVersion,
        minEpochsBetweenUpdates.toNumber(),
        false,
        false
      )
    )
      .to.be.revertedWithCustomError(
        fixture.dynamics,
        "InvalidAliceNetNodeHash"
      )
      .withArgs(
        newAliceNetVersion.binaryHash,
        "0xbc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"
      );
  });

  it("Should not be possible update AliceNet before minUpdate time", async () => {
    const newAliceNetVersion = {
      ...alicenetCurrentVersion,
    };
    newAliceNetVersion.major += 1;
    newAliceNetVersion.minor += 1;
    newAliceNetVersion.patch += 1;

    await expect(
      updateAliceNetNode(
        fixture,
        newAliceNetVersion,
        minEpochsBetweenUpdates.toNumber() - 1,
        false,
        false
      )
    )
      .to.be.revertedWithCustomError(fixture.dynamics, "InvalidScheduledDate")
      .withArgs(
        minEpochsBetweenUpdates.toNumber() - 1,
        minEpochsBetweenUpdates,
        maxEpochsBetweenUpdates
      );
  });

  it("Should not be possible update AliceNet before minUpdate time", async () => {
    const newAliceNetVersion = {
      ...alicenetCurrentVersion,
    };
    newAliceNetVersion.major += 1;
    newAliceNetVersion.minor += 1;
    newAliceNetVersion.patch += 1;

    await expect(
      updateAliceNetNode(
        fixture,
        newAliceNetVersion,
        maxEpochsBetweenUpdates.toNumber() + 1,
        false,
        false
      )
    )
      .to.be.revertedWithCustomError(fixture.dynamics, "InvalidScheduledDate")
      .withArgs(
        maxEpochsBetweenUpdates.toNumber() + 1,
        minEpochsBetweenUpdates,
        maxEpochsBetweenUpdates
      );
  });
});
