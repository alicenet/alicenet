import { BigNumber, BytesLike } from "ethers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import {
  BaseTokensFixture,
  Fixture,
  getContractAddressFromDeployedRawEvent,
  getFixture,
} from "../setup";
import { getState, showState } from "./setup";

const updateDistributionContract = async (
  fixture: Fixture | BaseTokensFixture,
  splits: number[]
) => {
  const transaction = await fixture.factory.deployCreate(
    (
      await ethers.getContractFactory("Distribution")
    ).getDeployTransaction(splits[0], splits[1], splits[2], splits[3])
      .data as BytesLike
  );
  await fixture.factory.upgradeProxy(
    ethers.utils.formatBytes32String("Distribution"),
    await getContractAddressFromDeployedRawEvent(transaction),
    "0x"
  );
};

const assertSplitsBalance = async (
  fixture: Fixture | BaseTokensFixture,
  splits: number[],
  distributable: BigNumber
) => {
  expect(
    await ethers.provider.getBalance(fixture.validatorStaking.address)
  ).to.be.equal(distributable.mul(splits[0]).div(1000));
  expect(
    await ethers.provider.getBalance(fixture.publicStaking.address)
  ).to.be.equal(distributable.mul(splits[1]).div(1000));
  expect(
    await ethers.provider.getBalance(fixture.liquidityProviderStaking.address)
  ).to.be.equal(distributable.mul(splits[2]).div(1000));
  expect(
    await ethers.provider.getBalance(fixture.foundation.address)
  ).to.be.equal(distributable.mul(splits[3]).div(1000));
};

const mintDistributeAndAssert = async (
  fixture: Fixture | BaseTokensFixture,
  splits: number[],
  ethIn: BigNumber,
  distributable: BigNumber
): Promise<BigNumber> => {
  await fixture.bToken.mint(0, { value: ethIn });
  distributable = distributable.add(await fixture.bToken.getYield());
  await fixture.bToken.distribute();
  expect(await fixture.bToken.getYield()).to.be.equals(0);
  await assertSplitsBalance(fixture, splits, distributable);
  return distributable;
};

describe("Testing BToken Distribution methods", async () => {
  let fixture: Fixture;
  const minBTokens = 0;
  const eth = 4;
  let ethIn: BigNumber;

  beforeEach(async function () {
    fixture = await getFixture();
    showState("Initial", await getState(fixture));
    ethIn = ethers.utils.parseEther(eth.toString());
  });

  it("Should not allow reentrancy on Distribution contract", async () => {
    const transaction = await fixture.factory.deployCreate(
      (
        await ethers.getContractFactory("ReentrantLoopDistributionMock")
      ).getDeployTransaction().data as BytesLike
    );
    await fixture.factory.upgradeProxy(
      ethers.utils.formatBytes32String("Distribution"),
      await getContractAddressFromDeployedRawEvent(transaction),
      "0x"
    );
    await fixture.bToken.mint(0, { value: ethIn });
    await expect(fixture.bToken.distribute()).to.be.revertedWithCustomError(
      fixture.bToken,
      "MutexLocked"
    );
  });

  it("Should not allow reentrancy on subCalls in the distribution contract", async () => {
    const transaction = await fixture.factory.deployCreate(
      (
        await ethers.getContractFactory("ReentrantLoopDistributionMock")
      ).getDeployTransaction().data as BytesLike
    );
    await fixture.factory.upgradeProxy(
      ethers.utils.formatBytes32String("Foundation"),
      await getContractAddressFromDeployedRawEvent(transaction),
      "0x"
    );
    await fixture.bToken.mint(0, { value: ethIn });
    await expect(fixture.bToken.distribute()).to.be.revertedWithCustomError(
      fixture.bToken,
      "MutexLocked"
    );
  });

  it("Should correctly distribute", async () => {
    const splits = [250, 250, 250, 250];
    await updateDistributionContract(fixture, splits);
    const distributable = await mintDistributeAndAssert(
      fixture,
      splits,
      ethIn,
      BigNumber.from(0)
    );
    await mintDistributeAndAssert(fixture, splits, ethIn, distributable);
  });

  it("Should correctly distribute big amount of eth", async () => {
    ethIn = ethers.utils.parseEther("70000000000");
    const splits = [250, 250, 250, 250];
    await updateDistributionContract(fixture, splits);
    const distributable = await mintDistributeAndAssert(
      fixture,
      splits,
      ethIn,
      BigNumber.from(0)
    );
    await mintDistributeAndAssert(fixture, splits, ethIn, distributable);
  });

  it("Should distribute without foundation", async () => {
    await fixture.bToken.mint(minBTokens, { value: ethIn });
    const splits = [350, 350, 300, 0];
    await updateDistributionContract(fixture, splits);
    const distributable = await mintDistributeAndAssert(
      fixture,
      splits,
      ethIn,
      BigNumber.from(0)
    );
    await mintDistributeAndAssert(fixture, splits, ethIn, distributable);
  });

  it("Should distribute without liquidityProviderStaking", async () => {
    const splits = [350, 350, 0, 300];
    await updateDistributionContract(fixture, splits);
    const distributable = await mintDistributeAndAssert(
      fixture,
      splits,
      ethIn,
      BigNumber.from(0)
    );
    await mintDistributeAndAssert(fixture, splits, ethIn, distributable);
  });

  it("Should distribute without publicStaking", async () => {
    const splits = [350, 0, 350, 300];
    await updateDistributionContract(fixture, splits);
    const distributable = await mintDistributeAndAssert(
      fixture,
      splits,
      ethIn,
      BigNumber.from(0)
    );
    await mintDistributeAndAssert(fixture, splits, ethIn, distributable);
  });

  it("Should distribute without validatorStaking", async () => {
    const splits = [0, 350, 350, 300];
    await updateDistributionContract(fixture, splits);
    const distributable = await mintDistributeAndAssert(
      fixture,
      splits,
      ethIn,
      BigNumber.from(0)
    );
    await mintDistributeAndAssert(fixture, splits, ethIn, distributable);
  });
});
