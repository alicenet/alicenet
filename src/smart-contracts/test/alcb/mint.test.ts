import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "./setup";

describe("Testing ALCB Minting methods", async () => {
  let fixture: Fixture;
  beforeEach(async function () {
    fixture = await loadFixture(getFixture);
  });

  it("Should mint alcbs equals to ether value sent", async () => {
    const amount1 = 100n;
    await fixture.alcb.mint({ value: amount1 });
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(amount1);
    expect(await fixture.alcb.totalSupply()).to.be.equal(amount1);
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(amount1);
    expect(await fixture.alcb.poolBalance()).to.be.equal(amount1);

    const amount2 = ethers.parseEther("1");
    await fixture.alcb.mint({ value: amount2 });
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(amount1 + amount2);
    expect(await fixture.alcb.totalSupply()).to.be.equal(amount1 + amount2);
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(amount1 + amount2);
    expect(await fixture.alcb.poolBalance()).to.be.equal(amount1 + amount2);
  });

  it("Should mintTo alcbs equals to ether value sent", async () => {
    const amount1 = 100n;
    await fixture.alcb.mintTo(await ethers.provider.getSigner(2), {
      value: amount1,
    });
    expect(
      await fixture.alcb.balanceOf(await ethers.provider.getSigner(2))
    ).to.be.equal(amount1);
    expect(await fixture.alcb.totalSupply()).to.be.equal(amount1);
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(amount1);
    expect(await fixture.alcb.poolBalance()).to.be.equal(amount1);

    const amount2 = ethers.parseEther("1");
    await fixture.alcb.mintTo(await ethers.provider.getSigner(3), {
      value: amount2,
    });
    expect(
      await fixture.alcb.balanceOf(await ethers.provider.getSigner(3))
    ).to.be.equal(amount2);
    expect(await fixture.alcb.totalSupply()).to.be.equal(amount1 + amount2);
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(amount1 + amount2);
    expect(await fixture.alcb.poolBalance()).to.be.equal(amount1 + amount2);
  });

  it("Should mint a big amount of eth to an address", async () => {
    const amount = 1000000000000000000n;
    await fixture.alcb.mintTo(await ethers.provider.getSigner(2), {
      value: amount,
    });
    expect(
      await fixture.alcb.balanceOf(await ethers.provider.getSigner(2))
    ).to.be.equal(amount);
    expect(await fixture.alcb.totalSupply()).to.be.equal(amount);
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(amount);
    expect(await fixture.alcb.poolBalance()).to.be.equal(amount);
  });

  it("Should mint a lot to multiple addresses", async () => {
    const amount = ethers.parseEther("1.5");
    let total = 0n;
    for (let i = 0; i < 20; i++) {
      await fixture.alcb.mintTo(await ethers.provider.getSigner(i), {
        value: amount,
      });
      expect(
        await fixture.alcb.balanceOf(await ethers.provider.getSigner(i))
      ).to.be.equal(amount);
      total += amount;
    }
    expect(await fixture.alcb.totalSupply()).to.be.equal(total);
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(total);
    expect(await fixture.alcb.poolBalance()).to.be.equal(total);
    expect(total).to.be.equal(ethers.parseEther("30"));
  });

  it("Should fail to mint to 0x0 address", async () => {
    await expect(
      fixture.alcb.mintTo(ethers.ZeroAddress, { value: 100n })
    ).to.be.revertedWithCustomError(fixture.alcb, "ERC20InvalidReceiver");
  });

  it("Should not mint if paused", async () => {
    await fixture.alcb.pauseMinting();
    await expect(
      fixture.alcb.mint({ value: 100n })
    ).to.be.revertedWithCustomError(fixture.alcb, "EnforcedPause");
  });
});
