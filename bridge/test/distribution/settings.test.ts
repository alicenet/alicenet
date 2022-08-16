import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { BigNumber } from "ethers";
import { ethers, expect } from "hardhat";
import { getState, showState } from "../bToken/setup";
import { Fixture, getFixture } from "../setup";

describe("Testing splits settings update", async () => {
  let fixture: Fixture;
  let admin: SignerWithAddress;

  beforeEach(async function () {
    fixture = await getFixture();
    [admin] = await ethers.getSigners();
    showState("Initial", await getState(fixture));
  });

  it("Should fail to set splits greater than one unit", async () => {
    await expect(
      (await ethers.getContractFactory("Distribution")).deploy(333, 333, 333, 2)
    ).to.be.revertedWithCustomError(fixture.distribution, `SplitValueSumError`);
  });

  it("Should fail to set all splits to 0", async () => {
    await expect(
      (await ethers.getContractFactory("Distribution")).deploy(0, 0, 0, 0)
    ).to.be.revertedWithCustomError(fixture.distribution, `SplitValueSumError`);
  });

  it("Should set some splits to 0 on a new deployment", async () => {
    const newDistribution = await (
      await ethers.getContractFactory("Distribution")
    ).deploy(0, 0, 1000, 0);
    await fixture.factory.upgradeProxy(
      ethers.utils.formatBytes32String("Distribution"),
      newDistribution.address,
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
    const newDistribution = await (
      await ethers.getContractFactory("Distribution")
    ).deploy(300, 300, 300, 100);
    await fixture.factory.upgradeProxy(
      ethers.utils.formatBytes32String("Distribution"),
      newDistribution.address,
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
