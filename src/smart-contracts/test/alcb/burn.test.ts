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

describe("Testing ALCB Burn methods", async () => {
  let fixture: Fixture;
  let publicStaking: PublicStaking;
  beforeEach(async function () {
    const superFixture = await loadFixture(deployFixture);
    fixture = superFixture.fixture;
    publicStaking = superFixture.publicStaking;
    await fixture.alcb.setOperationalCostFactor(50n);
    await fixture.alcb.setMarketSpreadFactor(250n);
  });

  it("Should burn alcb and distribute the correct amount", async () => {
    const balanceOwnerBeforeOperation = await ethers.provider.getBalance(
      await fixture.owner.getAddress()
    );
    const shareHolders = await fixture.alcb.getShareHolders();
    const shareHolder0Balance = await ethers.provider.getBalance(
      shareHolders[0].account
    );
    const shareHolder1Balance = await ethers.provider.getBalance(
      shareHolders[1].account
    );
    const rcpt = await (await fixture.alcb.burn(mintedAmount)).wait();
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(0n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(0n);
    expect(await fixture.alcb.poolBalance()).to.be.equal(0n);
    const expectedUserAdditionalBalance = (mintedAmount * 25n) / 100n;
    const expectedYieldAmount = mintedAmount - expectedUserAdditionalBalance;
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
      await ethers.provider.getBalance(await fixture.owner.getAddress())
    ).to.be.equal(
      expectedUserAdditionalBalance +
        balanceOwnerBeforeOperation -
        (rcpt?.gasUsed || 0n) * (rcpt?.gasPrice || 0n)
    );
  });

  it("Should burnTo alcb to an address and distribute the correct amount", async () => {
    const randomUser = ethers.Wallet.createRandom();
    const shareHolders = await fixture.alcb.getShareHolders();
    const shareHolder0Balance = await ethers.provider.getBalance(
      shareHolders[0].account
    );
    const shareHolder1Balance = await ethers.provider.getBalance(
      shareHolders[1].account
    );
    await fixture.alcb.burnTo(randomUser, mintedAmount);
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(0n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(0n);
    expect(await fixture.alcb.poolBalance()).to.be.equal(0n);
    const expectedUserAdditionalBalance = (mintedAmount * 25n) / 100n;
    const expectedYieldAmount = mintedAmount - expectedUserAdditionalBalance;
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
    expect(await ethers.provider.getBalance(randomUser)).to.be.equal(
      expectedUserAdditionalBalance
    );
  });

  it("Should burn alcb and distribute the correct amount to multiple share holders", async () => {
    await fixture.alcb.setShareHolders([
      {
        account: await (await ethers.provider.getSigner(3)).getAddress(),
        percentage: 101n,
        isMagicTransfer: false,
      },
      {
        account: await publicStaking.getAddress(),
        percentage: 667n,
        isMagicTransfer: true,
      },
      {
        account: await (await ethers.provider.getSigner(4)).getAddress(),
        percentage: 111n,
        isMagicTransfer: false,
      },
      {
        account: await (await ethers.provider.getSigner(5)).getAddress(),
        percentage: 121n,
        isMagicTransfer: false,
      },
    ]);
    await fixture.alcb.setMarketSpreadFactor(0);
    const shareHolders = await fixture.alcb.getShareHolders();
    const shareHoldersBalances = await Promise.all(
      shareHolders.map((shareHolder) =>
        ethers.provider.getBalance(shareHolder.account)
      )
    );
    await fixture.alcb.burn(333n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(mintedAmount - 333n);
    // the last of the share holders should be getting the remainder of the integer division,
    // therefore instead of receiving 40 if we would to compute normally, it should receive 42
    const expectedBalancesOffset = [33n, 222n, 36n, 42n];
    for (let i = 0; i < shareHolders.length; i++) {
      expect(
        await ethers.provider.getBalance(shareHolders[i].account)
      ).to.be.equal(
        shareHoldersBalances[i] + expectedBalancesOffset[i],
        "i=" + i
      );
    }
  });

  it("Should burn alcb and distribute no amount if the market spread is 100%", async () => {
    const shareHolders = await fixture.alcb.getShareHolders();
    const shareHolder0Balance = await ethers.provider.getBalance(
      shareHolders[0].account
    );
    const shareHolder1Balance = await ethers.provider.getBalance(
      shareHolders[1].account
    );
    await fixture.alcb.setMarketSpreadFactor(1000);
    const balanceOwnerBeforeOperation = await ethers.provider.getBalance(
      await fixture.owner.getAddress()
    );
    const rcpt = await (await fixture.alcb.burn(mintedAmount)).wait();
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
    ).to.be.equal(shareHolder0Balance);
    expect(
      await ethers.provider.getBalance(shareHolders[1].account)
    ).to.be.equal(shareHolder1Balance);
    expect(
      await ethers.provider.getBalance(await fixture.owner.getAddress())
    ).to.be.equal(
      mintedAmount +
        balanceOwnerBeforeOperation -
        (rcpt?.gasUsed || 0n) * (rcpt?.gasPrice || 0n)
    );
  });

  it("Should burn alcb bit by bit and distribute the correct amount", async () => {
    const balanceOwnerBeforeOperation = await ethers.provider.getBalance(
      await fixture.owner.getAddress()
    );
    const shareHolders = await fixture.alcb.getShareHolders();
    const shareHolder0Balance = await ethers.provider.getBalance(
      shareHolders[0].account
    );
    const shareHolder1Balance = await ethers.provider.getBalance(
      shareHolders[1].account
    );
    const burnedAmount = mintedAmount / 2n;
    const rcpt = await (await fixture.alcb.burn(burnedAmount)).wait();
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(mintedAmount / 2n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(mintedAmount / 2n);
    expect(await fixture.alcb.poolBalance()).to.be.equal(mintedAmount / 2n);
    const expectedUserAdditionalBalance = (burnedAmount * 25n) / 100n;
    const expectedYieldAmount = burnedAmount - expectedUserAdditionalBalance;
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
      await ethers.provider.getBalance(await fixture.owner.getAddress())
    ).to.be.equal(
      expectedUserAdditionalBalance +
        balanceOwnerBeforeOperation -
        (rcpt?.gasUsed || 0n) * (rcpt?.gasPrice || 0n)
    );
    const rcpt2 = await (await fixture.alcb.burn(mintedAmount / 2n)).wait();
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(0n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(0n);
    expect(await fixture.alcb.poolBalance()).to.be.equal(0n);
    const expectedUserAdditionalBalance2 = (burnedAmount * 25n) / 100n;
    const expectedYieldAmount2 = burnedAmount - expectedUserAdditionalBalance2;
    const expectedPublicStakingBalance2 = (expectedYieldAmount2 * 60n) / 100n;
    const expectedUserBalance2 = (expectedYieldAmount2 * 40n) / 100n;
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(0n);
    expect(
      await ethers.provider.getBalance(shareHolders[0].account)
    ).to.be.equal(
      expectedPublicStakingBalance2 +
        expectedPublicStakingBalance +
        shareHolder0Balance
    );
    expect(
      await ethers.provider.getBalance(shareHolders[1].account)
    ).to.be.equal(
      expectedUserBalance2 + expectedUserBalance + shareHolder1Balance
    );
    expect(
      await ethers.provider.getBalance(await fixture.owner.getAddress())
    ).to.be.equal(
      expectedUserAdditionalBalance +
        expectedUserAdditionalBalance2 +
        balanceOwnerBeforeOperation -
        (rcpt2?.gasUsed || 0n) * (rcpt2?.gasPrice || 0n) -
        (rcpt?.gasUsed || 0n) * (rcpt?.gasPrice || 0n)
    );
  });

  it("Should burn alcb and distribute the correct amount with 1 shareholder", async () => {
    const shareHolders = await fixture.alcb.getShareHolders();
    await fixture.alcb.setShareHolders([
      {
        account: await publicStaking.getAddress(),
        percentage: 1000n,
        isMagicTransfer: true,
      },
    ]);
    const balanceOwnerBeforeOperation = await ethers.provider.getBalance(
      await fixture.owner.getAddress()
    );
    const shareHolder0Balance = await ethers.provider.getBalance(
      shareHolders[0].account
    );
    const rcpt = await (await fixture.alcb.burn(mintedAmount)).wait();
    expect(
      await fixture.alcb.balanceOf(await fixture.owner.getAddress())
    ).to.be.equal(0n);
    expect(await fixture.alcb.totalSupply()).to.be.equal(0n);
    expect(await fixture.alcb.poolBalance()).to.be.equal(0n);
    const expectedUserAdditionalBalance = (mintedAmount * 25n) / 100n;
    const expectedYieldAmount = mintedAmount - expectedUserAdditionalBalance;
    const expectedPublicStakingBalance = (expectedYieldAmount * 100n) / 100n;
    expect(
      await ethers.provider.getBalance(await fixture.alcb.getAddress())
    ).to.be.equal(0n);
    expect(
      await ethers.provider.getBalance(shareHolders[0].account)
    ).to.be.equal(expectedPublicStakingBalance + shareHolder0Balance);
    expect(
      await ethers.provider.getBalance(await fixture.owner.getAddress())
    ).to.be.equal(
      expectedUserAdditionalBalance +
        balanceOwnerBeforeOperation -
        (rcpt?.gasUsed || 0n) * (rcpt?.gasPrice || 0n)
    );
  });

  it("Should not burn more alcbs than supply", async () => {
    await expect(
      fixture.alcb.burn(mintedAmount + 1n)
    ).to.be.revertedWithCustomError(fixture.alcb, "InvalidBurnAmount");
  });

  it("Should not burn more alcbs than on balance", async () => {
    await fixture.alcb.mintTo(await ethers.provider.getSigner(3), {
      value: 100n,
    });
    await expect(
      fixture.alcb.burn(mintedAmount + 1n)
    ).to.be.revertedWithCustomError(fixture.alcb, "ERC20InsufficientBalance");
  });

  it("Should burn 0 tokens", async () => {
    await fixture.alcb.burn(0n);
  });

  it("Should burn tokens if the contract is paused", async () => {
    await fixture.alcb.pauseMinting();
    await fixture.alcb.burn(mintedAmount);
  });

  it("Should not burn others people alcbs", async () => {
    await expect(
      fixture.alcb
        .connect(await ethers.provider.getSigner(3))
        .burn(mintedAmount)
    ).to.be.revertedWithCustomError(fixture.alcb, "ERC20InsufficientBalance");
  });
});
