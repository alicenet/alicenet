import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers } from "hardhat";
import { Dynamics } from "../../typechain-types";
import { DynamicValuesStruct } from "../../typechain-types/contracts/Dynamics.sol/Dynamics";
import { expect } from "../chai-setup";
import { factoryCallAny, Fixture, getFixture } from "../setup";

describe("Testing Dynamics methods", async () => {
  let admin: SignerWithAddress;
  let user: SignerWithAddress;
  let dynamics: Dynamics;
  let fixture: Fixture;
  let emptyArray = new Array();

  const minEpochsBetweenUpdates = BigNumber.from(2);
  const maxEpochsBetweenUpdates = BigNumber.from(336);
  const Version = {
    V1: BigNumber.from(1),
  };

  let dynamicValues: DynamicValuesStruct =
  {
    encoderVersion: 0,
    proposalTimeout: 4000,
    preVoteTimeout: 3000,
    preCommitTimeout: 3000,
    maxBlockSize: BigNumber.from(3000000),
    dataStoreFee: BigNumber.from(0),
    valueStoreFee: BigNumber.from(0),
    minScaledTransactionFee: BigNumber.from(0)
  }

  beforeEach(async function () {
    fixture = await getFixture(false, true, false);
    const signers = await ethers.getSigners();
    [admin, user] = signers;
    dynamics = fixture.dynamics;
  });

  it("Should get initial Configuration", async () => {
    const configuration = await dynamics.getConfiguration();
    expect(configuration[0]).to.be.equal(minEpochsBetweenUpdates);
    expect(configuration[1]).to.be.equal(maxEpochsBetweenUpdates);
  });

  it("Should not change dynamic values if not impersonating factory", async () => {
    await expect(dynamics.changeDynamicValues(BigNumber.from(1), dynamicValues)).to.be.revertedWith("2000");
  });

  it("Should not change dynamic values to a epoch lesser than minEpochsBetweenUpdates if impersonating factory", async () => {
    await expect(factoryCallAny(fixture.factory, dynamics, "changeDynamicValues", [minEpochsBetweenUpdates.sub(1), dynamicValues])).to.be.revertedWith("");
  });

  it("Should not change dynamic values to a epoch greater than maxEpochsBetweenUpdates if impersonating factory", async () => {
    await expect(factoryCallAny(fixture.factory, dynamics, "changeDynamicValues", [maxEpochsBetweenUpdates.add(1), dynamicValues])).to.be.revertedWith("");
  });

  it("Should change dynamic values to a valid epoch if impersonating factory", async () => {
    dynamicValues.valueStoreFee = BigNumber.from(1)
    await factoryCallAny(fixture.factory, dynamics, "changeDynamicValues", [BigNumber.from(3), dynamicValues]);
    await factoryCallAny(fixture.factory, fixture.snapshots, "snapshot", [emptyArray, emptyArray]);
    expect((await dynamics.getLatestDynamicValues())[6]).to.be.equal(BigNumber.from(0))
    await factoryCallAny(fixture.factory, fixture.snapshots, "snapshot", [emptyArray, emptyArray]);
    expect((await dynamics.getLatestDynamicValues())[6]).to.be.equal(BigNumber.from(0))
    await factoryCallAny(fixture.factory, fixture.snapshots, "snapshot", [emptyArray, emptyArray]);
    expect((await dynamics.getLatestDynamicValues())[6]).to.be.equal(BigNumber.from(1))
  });

  it("Should get latest dynamic values ", async () => {
    const dv = await dynamics.getLatestDynamicValues();
    expect(dv).to.deep.equal(dynamicValues)
  });

});
