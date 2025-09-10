import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { Fixture, getFixture } from "./setup";
import { ethers } from "hardhat";
import { expect } from "chai";

describe("Testing ALCB Utils methods", async () => {
  let fixture: Fixture;
  beforeEach(async function () {
    fixture = await loadFixture(getFixture);
  });

  it("Only owner should be able to pause and unpause minting", async () => {
    const notOwner = await ethers.provider.getSigner(2);
    await expect(fixture.alcb.connect(notOwner).pauseMinting())
      .to.be.revertedWithCustomError(fixture.alcb, "OwnableUnauthorizedAccount")
      .withArgs(await notOwner.getAddress());
    await expect(fixture.alcb.connect(notOwner).unpauseMinting())
      .to.be.revertedWithCustomError(fixture.alcb, "OwnableUnauthorizedAccount")
      .withArgs(await notOwner.getAddress());
    await expect(fixture.alcb.connect(fixture.owner).pauseMinting()).to.emit(
      fixture.alcb,
      "Paused"
    );
    expect(await fixture.alcb.paused()).to.be.equal(true);
    await expect(fixture.alcb.connect(fixture.owner).unpauseMinting()).to.emit(
      fixture.alcb,
      "Unpaused"
    );
    expect(await fixture.alcb.paused()).to.be.equal(false);
  });

  it("Only owner should be able to set market spread factor", async () => {
    const notOwner = await ethers.provider.getSigner(2);
    await expect(fixture.alcb.connect(notOwner).setMarketSpreadFactor(100))
      .to.be.revertedWithCustomError(fixture.alcb, "OwnableUnauthorizedAccount")
      .withArgs(await notOwner.getAddress());
    await fixture.alcb.connect(fixture.owner).setMarketSpreadFactor(100);
    expect(await fixture.alcb.marketSpreadFactor()).to.be.equal(100);
    await fixture.alcb.connect(fixture.owner).setMarketSpreadFactor(1000);
    expect(await fixture.alcb.marketSpreadFactor()).to.be.equal(1000);
    await fixture.alcb.connect(fixture.owner).setMarketSpreadFactor(0);
    expect(await fixture.alcb.marketSpreadFactor()).to.be.equal(0);
  });

  it("Market spread factor shouldn't be greater than scale factor", async () => {
    await expect(
      fixture.alcb.connect(fixture.owner).setMarketSpreadFactor(1001)
    )
      .to.be.revertedWithCustomError(fixture.alcb, "MarketSpreadFactorError")
      .withArgs(1001, 1000);
  });

  it("Only owner should be able to set operational factor", async () => {
    const notOwner = await ethers.provider.getSigner(2);
    await expect(fixture.alcb.connect(notOwner).setOperationalCostFactor(100))
      .to.be.revertedWithCustomError(fixture.alcb, "OwnableUnauthorizedAccount")
      .withArgs(await notOwner.getAddress());
    await fixture.alcb.connect(fixture.owner).setOperationalCostFactor(100);
    expect(await fixture.alcb.operationalCostFactor()).to.be.equal(100);
    await fixture.alcb.connect(fixture.owner).setOperationalCostFactor(1000);
    expect(await fixture.alcb.operationalCostFactor()).to.be.equal(1000);
    await fixture.alcb.connect(fixture.owner).setOperationalCostFactor(0);
    expect(await fixture.alcb.operationalCostFactor()).to.be.equal(0);
  });

  it("Operational factor shouldn't be greater than scale factor", async () => {
    await expect(
      fixture.alcb.connect(fixture.owner).setOperationalCostFactor(1001)
    )
      .to.be.revertedWithCustomError(fixture.alcb, "OperationalFactorError")
      .withArgs(1001, 1000);
  });

  it("Only owner should be able to set share holders", async () => {
    const notOwner = await ethers.provider.getSigner(2);
    const shareHolders = [
      {
        account: await notOwner.getAddress(),
        percentage: 1000n,
        isMagicTransfer: false,
      },
    ];
    await expect(fixture.alcb.connect(notOwner).setShareHolders(shareHolders))
      .to.be.revertedWithCustomError(fixture.alcb, "OwnableUnauthorizedAccount")
      .withArgs(await notOwner.getAddress());
    await fixture.alcb.connect(fixture.owner).setShareHolders(shareHolders);
    const contractShareHolders = await fixture.alcb.getShareHolders();
    contractShareHolders.forEach((shareHolder, index) => {
      expect(shareHolder.account).to.be.equal(shareHolders[index].account);
      expect(shareHolder.percentage).to.be.equal(
        shareHolders[index].percentage
      );
      expect(shareHolder.isMagicTransfer).to.be.equal(
        shareHolders[index].isMagicTransfer
      );
    });
    const shareHolders2 = [
      {
        account: await notOwner.getAddress(),
        percentage: 900n,
        isMagicTransfer: false,
      },
      {
        account: await notOwner.getAddress(),
        percentage: 100n,
        isMagicTransfer: false,
      },
    ];
    await fixture.alcb.connect(fixture.owner).setShareHolders(shareHolders2);
    const contractShareHolders2 = await fixture.alcb.getShareHolders();
    expect(await fixture.alcb.getShareHolderCount()).to.be.equal(2);
    contractShareHolders2.forEach((shareHolder, index) => {
      expect(shareHolder.account).to.be.equal(shareHolders2[index].account);
      expect(shareHolder.percentage).to.be.equal(
        shareHolders2[index].percentage
      );
      expect(shareHolder.isMagicTransfer).to.be.equal(
        shareHolders2[index].isMagicTransfer
      );
    });
  });

  it("Share holders percentages should not pass 1000", async () => {
    await expect(
      fixture.alcb.connect(fixture.owner).setShareHolders([
        {
          account: await (await ethers.provider.getSigner(2)).getAddress(),
          percentage: 600n,
          isMagicTransfer: false,
        },
        {
          account: await (await ethers.provider.getSigner(3)).getAddress(),
          percentage: 500n,
          isMagicTransfer: false,
        },
      ])
    )
      .to.be.revertedWithCustomError(
        fixture.alcb,
        "ShareHoldersPercentageSumError"
      )
      .withArgs(1100, 1000);
    await expect(
      fixture.alcb.connect(fixture.owner).setShareHolders([
        {
          account: await (await ethers.provider.getSigner(2)).getAddress(),
          percentage: 1001n,
          isMagicTransfer: false,
        },
      ])
    )
      .to.be.revertedWithCustomError(
        fixture.alcb,
        "ShareHoldersPercentageSumError"
      )
      .withArgs(1001, 1000);
  });

  it("Share holders percentages should not be 0", async () => {
    await expect(
      fixture.alcb.connect(fixture.owner).setShareHolders([
        {
          account: await (await ethers.provider.getSigner(2)).getAddress(),
          percentage: 0n,
          isMagicTransfer: false,
        },
        {
          account: await (await ethers.provider.getSigner(3)).getAddress(),
          percentage: 1000n,
          isMagicTransfer: false,
        },
      ])
    ).to.be.revertedWithCustomError(fixture.alcb, "PercentagesCannotBeZero");
  });

  it("Share holders address should not be 0", async () => {
    await expect(
      fixture.alcb.connect(fixture.owner).setShareHolders([
        {
          account: await (await ethers.provider.getSigner(2)).getAddress(),
          percentage: 1n,
          isMagicTransfer: false,
        },
        {
          account: ethers.ZeroAddress,
          percentage: 999n,
          isMagicTransfer: false,
        },
      ])
    ).to.be.revertedWithCustomError(
      fixture.alcb,
      "CannotTransferToZeroAddress"
    );
  });

  it("Only owner should be able to set operational wallet", async () => {
    const notOwner = await ethers.provider.getSigner(2);
    await expect(
      fixture.alcb
        .connect(notOwner)
        .setOperationalWallet(await notOwner.getAddress())
    )
      .to.be.revertedWithCustomError(fixture.alcb, "OwnableUnauthorizedAccount")
      .withArgs(await notOwner.getAddress());
    await fixture.alcb
      .connect(fixture.owner)
      .setOperationalWallet(await notOwner.getAddress());
    expect(await fixture.alcb.operationalWallet()).to.be.equal(
      await notOwner.getAddress()
    );
  });
});
