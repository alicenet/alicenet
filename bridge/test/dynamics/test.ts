import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import {
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
    maxBlockSize: BigNumber.from(3000000),
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

  beforeEach(async function () {
    fixture = await getFixture(false, true, false);
    const signers = await ethers.getSigners();
    [admin, user] = signers;
  });

  it("Should get initial Configuration", async () => {
    const configuration =
      (await fixture.dynamics.getConfiguration()) as ConfigurationStruct;
    expect(configuration.minEpochsBetweenUpdates).to.be.equal(
      minEpochsBetweenUpdates
    );
    expect(configuration.maxEpochsBetweenUpdates).to.be.equal(
      maxEpochsBetweenUpdates
    );
  });

  it("Should not change dynamic values if not impersonating factory", async () => {
    await expect(
      fixture.dynamics.changeDynamicValues(futureEpoch, currentDynamicValues)
    )
      .to.be.revertedWithCustomError(fixture.dynamics, "OnlyFactory")
      .withArgs(admin.address, fixture.factory.address);
  });

  it("Should not change dynamic values to a epoch lesser than minEpochsBetweenUpdates", async () => {
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

  it("Should not change dynamic values to a epoch greater than maxEpochsBetweenUpdates", async () => {
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

  it.skip("Should get latest dynamic values", async () => {
    const latestDynamicValues =
      (await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct;
    expect(
      (await fixture.dynamics.getLatestDynamicValues()) as DynamicValuesStruct
    ).to.be.deep.equal(currentDynamicValues);
  });

  it("Should change dynamic values to a valid epoch if impersonating factory", async () => {
    await factoryCallAny(
      fixture.factory,
      fixture.dynamics,
      "changeDynamicValues",
      [futureEpoch, currentDynamicValues]
    );
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
    ).to.be.equal(currentDynamicValues.valueStoreFee);
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
    // TODO: add event checking
  });
});
