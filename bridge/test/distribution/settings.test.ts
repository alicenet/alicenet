import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber, BytesLike, ContractTransaction } from "ethers";
import { ethers, expect } from "hardhat";
import { AliceNetFactory } from "../../typechain-types";
import { getState, showState } from "../bToken/setup";
import {
  Fixture,
  getContractAddressFromDeployedRawEvent,
  getFixture,
} from "../setup";

export const deployDistributionViaFactory = async (
  factory: AliceNetFactory,
  validatorStakingSplit_: number,
  publicStakingSplit_: number,
  liquidityProviderStakingSplit_: number,
  protocolFeeSplit_: number
) => {
  const deployData = (
    await ethers.getContractFactory("Distribution")
  ).getDeployTransaction(
    validatorStakingSplit_,
    publicStakingSplit_,
    liquidityProviderStakingSplit_,
    protocolFeeSplit_
  ).data as BytesLike;
  return factory.deployCreate(deployData);
};

export const getDistributionAddress = async (
  transaction: ContractTransaction
) => {
  const distributionAddress = await getContractAddressFromDeployedRawEvent(
    transaction
  );
  return distributionAddress;
};

describe("Testing splits settings update", async () => {
  let fixture: Fixture;
  let admin: SignerWithAddress;

  beforeEach(async function () {
    fixture = await loadFixture(getFixture);
    [admin] = await ethers.getSigners();
    showState("Initial", await getState(fixture));
  });

  it("Should fail to set splits greater than one unit", async () => {
    await expect(
      deployDistributionViaFactory(fixture.factory, 333, 333, 333, 2)
    ).to.be.reverted;
  });

  it("Should fail to set all splits to 0", async () => {
    await expect(deployDistributionViaFactory(fixture.factory, 0, 0, 0, 0)).to
      .be.reverted;
  });

  it("Should set some splits to 0 on a new deployment", async () => {
    const newDistribution = await getDistributionAddress(
      await deployDistributionViaFactory(fixture.factory, 0, 0, 1000, 0)
    );
    await fixture.factory.upgradeProxy(
      ethers.utils.formatBytes32String("Distribution"),
      newDistribution,
      "0x"
    );
    expect(await fixture.distribution.getSplits()).to.be.deep.equals([
      BigNumber.from(0),
      BigNumber.from(0),
      BigNumber.from(1000),
      BigNumber.from(0),
    ]);
  });

  it("Should correctly set the splits on a new deployment", async () => {
    expect(await fixture.distribution.getSplits()).to.be.deep.equals([
      BigNumber.from(332),
      BigNumber.from(332),
      BigNumber.from(332),
      BigNumber.from(4),
    ]);
    const newDistribution = await getDistributionAddress(
      await deployDistributionViaFactory(fixture.factory, 300, 300, 300, 100)
    );
    await fixture.factory.upgradeProxy(
      ethers.utils.formatBytes32String("Distribution"),
      newDistribution,
      "0x"
    );
    expect(await fixture.distribution.getSplits()).to.be.deep.equals([
      BigNumber.from(300),
      BigNumber.from(300),
      BigNumber.from(300),
      BigNumber.from(100),
    ]);
  });

  it("Only BToken should be able to distribute", async () => {
    await expect(
      fixture.distribution.connect(admin).depositEth(42, { value: 10000 })
    )
      .to.revertedWithCustomError(fixture.distribution, "OnlyBToken")
      .withArgs(admin.address, fixture.bToken.address);
    const rcpt = await (await fixture.bToken.distribute()).wait();
    expect(rcpt.status).to.be.equals(1);
  });

  it("Should not be able to send ethereum", async () => {
    await expect(
      admin.sendTransaction({ to: fixture.distribution.address, value: 10000 })
    ).to.reverted;
  });
});
