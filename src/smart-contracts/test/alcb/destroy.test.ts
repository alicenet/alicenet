import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { ethers } from "hardhat";
import { expect } from "../chai-setup";
import { Fixture, getFixture } from "./setup";
import { deployFactoryAndBaseTokens } from "../legacy/setup";
import { PublicStaking } from "../../typechain-types";

const mintedAmount = ethers.parseEther("1000");

async function deployFixture() {
  const { publicStaking } = await deployFactoryAndBaseTokens();
  const fixture = await getFixture();
  await fixture.alcb.mint({ value: mintedAmount });
  await fixture.alcb.setShareHolders([
    {
      account: await publicStaking.getAddress(),
      percentage: 600n,
      isMagicTransfer: true,
    },
    {
      account: await (await ethers.provider.getSigner(3)).getAddress(),
      percentage: 400n,
      isMagicTransfer: false,
    },
  ]);
  await fixture.alcb.setOperationalWallet(
    await fixture.operationalWallet.getAddress()
  );
  return { fixture, publicStaking };
}

describe("Testing ALCB Destroy methods", async () => {
  let fixture: Fixture;
  let publicStaking: PublicStaking;
  beforeEach(async function () {
    const superFixture = await loadFixture(deployFixture);
    fixture = superFixture.fixture;
    publicStaking = superFixture.publicStaking;
    await fixture.alcb.setOperationalCostFactor(50n);
  });

  it("Should destroy alcb and distribute the correct amount", async () => {
    const balanceBeforeOperation = await ethers.provider.getBalance(
      await fixture.alcb.operationalWallet()
    );
    const shareHolders = await fixture.alcb.getShareHolders();
    const shareHolder0Balance = await ethers.provider.getBalance(
      shareHolders[0].account
    );
    const shareHolder1Balance = await ethers.provider.getBalance(
      shareHolders[1].account
    );
    await fixture.alcb.destroyTokens(mintedAmount);
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(0n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(0n);
    expect(await fixture.alcb.poolBalance()).to.be.equal(0n);
    const expectedOperationalWalletBalance = (mintedAmount * 5n) / 100n;
    const expectedYieldAmount = mintedAmount - expectedOperationalWalletBalance;
    const expectedPublicStakingBalance = (expectedYieldAmount * 60n) / 100n;
    const expectedUserBalance = (expectedYieldAmount * 40n) / 100n;
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(0n);
    expect(
      await ethers.provider.getBalance(shareHolders[0].account)
    ).to.be.equal(expectedPublicStakingBalance + shareHolder0Balance);
    expect(
      await ethers.provider.getBalance(shareHolders[1].account)
    ).to.be.equal(expectedUserBalance + shareHolder1Balance);
    expect(
      await ethers.provider.getBalance(
        await fixture.operationalWallet.getAddress()
      )
    ).to.be.equal(expectedOperationalWalletBalance + balanceBeforeOperation);
  });

  it("Should destroy alcb bit by bit and distribute the correct amount", async () => {
    const balanceBeforeOperation = await ethers.provider.getBalance(
      await fixture.alcb.operationalWallet()
    );
    const shareHolders = await fixture.alcb.getShareHolders();
    const shareHolder0Balance = await ethers.provider.getBalance(
      shareHolders[0].account
    );
    const shareHolder1Balance = await ethers.provider.getBalance(
      shareHolders[1].account
    );
    const amountToDestroy = mintedAmount / 2n;
    await fixture.alcb.destroyTokens(amountToDestroy);
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(mintedAmount / 2n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(mintedAmount / 2n);
    expect(await fixture.alcb.poolBalance()).to.be.equal(mintedAmount / 2n);
    const expectedOperationalWalletBalance = (mintedAmount * 5n) / 100n / 2n;
    const expectedYieldAmount =
      mintedAmount / 2n - expectedOperationalWalletBalance;
    const expectedPublicStakingBalance = (expectedYieldAmount * 60n) / 100n;
    const expectedUserBalance = (expectedYieldAmount * 40n) / 100n;
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(mintedAmount / 2n);
    expect(
      await ethers.provider.getBalance(shareHolders[0].account)
    ).to.be.equal(expectedPublicStakingBalance + shareHolder0Balance);
    expect(
      await ethers.provider.getBalance(shareHolders[1].account)
    ).to.be.equal(expectedUserBalance + shareHolder1Balance);
    expect(
      await ethers.provider.getBalance(
        await fixture.operationalWallet.getAddress()
      )
    ).to.be.equal(expectedOperationalWalletBalance + balanceBeforeOperation);
    await fixture.alcb.destroyTokens(amountToDestroy);
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(0n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(0n);
    expect(await fixture.alcb.poolBalance()).to.be.equal(0n);
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(0n);
    expect(
      await ethers.provider.getBalance(shareHolders[0].account)
    ).to.be.equal(expectedPublicStakingBalance * 2n + shareHolder0Balance);
    expect(
      await ethers.provider.getBalance(shareHolders[1].account)
    ).to.be.equal(expectedUserBalance * 2n + shareHolder1Balance);
    expect(
      await ethers.provider.getBalance(
        await fixture.operationalWallet.getAddress()
      )
    ).to.be.equal(
      expectedOperationalWalletBalance * 2n + balanceBeforeOperation
    );
  });

  it("Should destroy alcb and distribute the correct amount with 1 shareholder", async () => {
    await fixture.alcb.setShareHolders([
      {
        account: await publicStaking.getAddress(),
        percentage: 1000n,
        isMagicTransfer: true,
      },
    ]);
    const balanceBeforeOperation = await ethers.provider.getBalance(
      await fixture.alcb.operationalWallet()
    );
    await fixture.alcb.destroyTokens(mintedAmount);
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(0n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(0n);
    expect(await fixture.alcb.poolBalance()).to.be.equal(0n);
    const expectedOperationalWalletBalance = (mintedAmount * 5n) / 100n;
    const expectedYieldAmount = mintedAmount - expectedOperationalWalletBalance;
    const expectedPublicStakingBalance = expectedYieldAmount;
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(0n);
    expect(
      await ethers.provider.getBalance(await publicStaking.getAddress())
    ).to.be.equal(expectedPublicStakingBalance);
    expect(
      await ethers.provider.getBalance(
        await fixture.operationalWallet.getAddress()
      )
    ).to.be.equal(expectedOperationalWalletBalance + balanceBeforeOperation);
  });

  it("Should not destroy more alcbs than supply", async () => {
    await expect(
      fixture.alcb.destroyTokens(mintedAmount + 1n)
    ).to.be.revertedWithCustomError(fixture.alcb, "InvalidBurnAmount");
  });

  it("Should not destroy more alcbs than on balance", async () => {
    await fixture.alcb.mintTo(await ethers.provider.getSigner(3), {
      value: 100n,
    });
    await expect(
      fixture.alcb.destroyTokens(mintedAmount + 1n)
    ).to.be.revertedWithCustomError(fixture.alcb, "ERC20InsufficientBalance");
  });

  it("Should destroy 0 tokens", async () => {
    await fixture.alcb.destroyTokens(0n);
  });

  it("Should destroy tokens if the contract is paused", async () => {
    await fixture.alcb.pauseMinting();
    await fixture.alcb.destroyTokens(mintedAmount);
  });

  it("Should not destroy others people alcbs", async () => {
    await expect(
      fixture.alcb
        .connect(await ethers.provider.getSigner(3))
        .destroyTokens(mintedAmount)
    ).to.be.revertedWithCustomError(fixture.alcb, "ERC20InsufficientBalance");
  });
});
