import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { expect } from "chai";
import { ethers } from "hardhat";
import { BaseTokensFixture, getBaseTokensFixture } from "../../setup";

describe("PublicStaking: Reentrancy tests", async () => {
  let fixture: BaseTokensFixture;

  beforeEach(async function () {
    fixture = await loadFixture(getBaseTokensFixture);
  });

  it("Should not allow reentrancy in the CollectEth", async function () {
    const ReentrantLoopEthCollectorAccount = await ethers.getContractFactory(
      "ReentrantLoopEthCollectorAccount"
    );
    const reentrantLoopEthCollectorAccount =
      await ReentrantLoopEthCollectorAccount.deploy();
    await reentrantLoopEthCollectorAccount.setTokens(
      await fixture.alca.getAddress(),
      await fixture.publicStaking.getAddress()
    );
    await fixture.alca.transfer(
      await reentrantLoopEthCollectorAccount.getAddress(),
      1000
    );
    await reentrantLoopEthCollectorAccount.approve(
      await fixture.publicStaking.getAddress(),
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks
    await reentrantLoopEthCollectorAccount.mint(1000);

    // mint to another user with the same amount
    await fixture.alca.approve(await fixture.publicStaking.getAddress(), 1000);
    await fixture.publicStaking.mint(1000);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.parseEther("1000"),
    });
    const tokenID = 1;
    const balanceBefore = await ethers.provider.getBalance(
      await reentrantLoopEthCollectorAccount.getAddress()
    );
    await reentrantLoopEthCollectorAccount.setTokenID(tokenID);
    // The collectEth reentrancy should only be called once, since the second time all the profit was
    // already transferred and from _safeTransferEth function we just return when the amount is 0
    await reentrantLoopEthCollectorAccount.collectEth(tokenID);
    // the balance of the dishonest user should only have increased by 500 ether (the right amount). If
    // this is not true it means that there was an exploit!
    expect(
      await ethers.provider.getBalance(
        await reentrantLoopEthCollectorAccount.getAddress()
      )
    ).to.be.equals(
      balanceBefore + ethers.parseEther("500"),
      "Balance after reentracy doesn't match!"
    );
  });

  it("Should not allow reentrancy that's is finite in the CollectEth", async function () {
    const ReentrantFiniteEthCollectorAccount = await ethers.getContractFactory(
      "ReentrantFiniteEthCollectorAccount"
    );
    const reentrantFiniteEthCollectorAccount =
      await ReentrantFiniteEthCollectorAccount.deploy();
    await reentrantFiniteEthCollectorAccount.setTokens(
      await fixture.alca.getAddress(),
      await fixture.publicStaking.getAddress()
    );
    await fixture.alca.transfer(
      await reentrantFiniteEthCollectorAccount.getAddress(),
      1000
    );
    await reentrantFiniteEthCollectorAccount.approve(
      await fixture.publicStaking.getAddress(),
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks
    await reentrantFiniteEthCollectorAccount.mint(1000);

    // mint to another user with the same amount
    await fixture.alca.approve(await fixture.publicStaking.getAddress(), 1000);
    await fixture.publicStaking.mint(1000);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.parseEther("1000"),
    });
    const tokenID = 1;
    const balanceBefore = await ethers.provider.getBalance(
      await reentrantFiniteEthCollectorAccount.getAddress()
    );
    await reentrantFiniteEthCollectorAccount.setTokenID(tokenID);
    // The collectEth reentrancy should only be called once, since the second time all the profit was
    // already transferred and from _safeTransferEth function we just return when the amount is 0
    await reentrantFiniteEthCollectorAccount.collectEth(tokenID);
    // the balance of the dishonest user should only have increased by 500 ether (the right amount). If
    // this is not true it means that there was an exploit!
    expect(
      await ethers.provider.getBalance(
        await reentrantFiniteEthCollectorAccount.getAddress()
      )
    ).to.be.equals(
      balanceBefore + ethers.parseEther("500"),
      "Balance after reentracy doesn't match!"
    );
  });

  it("Should not allow reentrancy when burning a position with finite loop", async function () {
    const ReentrantFiniteBurnAccount = await ethers.getContractFactory(
      "ReentrantFiniteBurnAccount"
    );
    const reentrantFiniteBurnAccount =
      await ReentrantFiniteBurnAccount.deploy();
    await reentrantFiniteBurnAccount.setTokens(
      await fixture.alca.getAddress(),
      await fixture.publicStaking.getAddress()
    );
    await fixture.alca.transfer(
      await reentrantFiniteBurnAccount.getAddress(),
      1000
    );
    await reentrantFiniteBurnAccount.approve(
      await fixture.publicStaking.getAddress(),
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks
    await reentrantFiniteBurnAccount.mint(1000);

    // mint to another user with the same amount
    await fixture.alca.approve(await fixture.publicStaking.getAddress(), 1000);
    await fixture.publicStaking.mint(1000);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.parseEther("1000"),
    });
    const tokenID = 1;
    await reentrantFiniteBurnAccount.setTokenID(tokenID);
    const reentrantFiniteBurnAccountAddress =
      await reentrantFiniteBurnAccount.getAddress();
    const expectedEthTransferAmount = BigInt(ethers.parseEther("500"));
    await expect(reentrantFiniteBurnAccount.burn(tokenID))
      .to.be.revertedWithCustomError(fixture.publicStaking, "EthTransferFailed")
      .withArgs(
        await fixture.publicStaking.getAddress(),
        reentrantFiniteBurnAccountAddress,
        expectedEthTransferAmount
      );
  });

  it("Should not allow burn reentrancy after transferring a NFT position", async function () {
    const ReentrantLoopBurnERC721ReceiverAccount =
      await ethers.getContractFactory("ReentrantLoopBurnERC721ReceiverAccount");
    const reentrantLoopBurnERC721ReceiverAccount =
      await ReentrantLoopBurnERC721ReceiverAccount.deploy();
    await reentrantLoopBurnERC721ReceiverAccount.setTokens(
      await fixture.alca.getAddress(),
      await fixture.publicStaking.getAddress()
    );
    await fixture.alca.transfer(
      await reentrantLoopBurnERC721ReceiverAccount.getAddress(),
      1000
    );
    await reentrantLoopBurnERC721ReceiverAccount.approve(
      await fixture.publicStaking.getAddress(),
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks

    // mint to another user with the same amount
    await fixture.alca.approve(await fixture.publicStaking.getAddress(), 1000);
    await fixture.publicStaking.mint(1000);
    await reentrantLoopBurnERC721ReceiverAccount.mint(1000);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.parseEther("1000"),
    });
    const tokenID = 1;

    const [adminSigner] = await ethers.getSigners();

    const expectedEthTransferAmount = BigInt(ethers.parseEther("500"));
    await expect(
      fixture.publicStaking["safeTransferFrom(address,address,uint256)"](
        adminSigner.address,
        await reentrantLoopBurnERC721ReceiverAccount.getAddress(),
        tokenID
      )
    )
      .to.be.revertedWithCustomError(fixture.publicStaking, "EthTransferFailed")
      .withArgs(
        await fixture.publicStaking.getAddress(),
        await reentrantLoopBurnERC721ReceiverAccount.getAddress(),
        expectedEthTransferAmount
      );
  });

  it("Should not allow burn reentrancy after transferring a NFT position finite loop", async function () {
    const ReentrantFiniteBurnERC721ReceiverAccount =
      await ethers.getContractFactory(
        "ReentrantFiniteBurnERC721ReceiverAccount"
      );
    const reentrantFiniteBurnERC721ReceiverAccount =
      await ReentrantFiniteBurnERC721ReceiverAccount.deploy();
    await reentrantFiniteBurnERC721ReceiverAccount.setTokens(
      await fixture.alca.getAddress(),
      await fixture.publicStaking.getAddress()
    );
    await fixture.alca.transfer(
      await reentrantFiniteBurnERC721ReceiverAccount.getAddress(),
      1000
    );
    await reentrantFiniteBurnERC721ReceiverAccount.approve(
      await fixture.publicStaking.getAddress(),
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks

    // mint to another user with the same amount
    await fixture.alca.approve(await fixture.publicStaking.getAddress(), 1000);
    await fixture.publicStaking.mint(1000);
    await reentrantFiniteBurnERC721ReceiverAccount.mint(1000);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.parseEther("1000"),
    });
    const tokenID = 1;

    const [adminSigner] = await ethers.getSigners();

    const expectedEthTransferAmount = BigInt(ethers.parseEther("500"));
    await expect(
      fixture.publicStaking["safeTransferFrom(address,address,uint256)"](
        adminSigner.address,
        await reentrantFiniteBurnERC721ReceiverAccount.getAddress(),
        tokenID
      )
    )
      .to.be.revertedWithCustomError(fixture.publicStaking, "EthTransferFailed")
      .withArgs(
        await fixture.publicStaking.getAddress(),
        await reentrantFiniteBurnERC721ReceiverAccount.getAddress(),
        expectedEthTransferAmount
      );
  });

  it("Should not allow collect reentrancy after transferring a NFT position", async function () {
    const ReentrantLoopCollectEthERC721ReceiverAccount =
      await ethers.getContractFactory(
        "ReentrantLoopCollectEthERC721ReceiverAccount"
      );
    const reentrantLoopCollectEthERC721ReceiverAccount =
      await ReentrantLoopCollectEthERC721ReceiverAccount.deploy();
    await reentrantLoopCollectEthERC721ReceiverAccount.setTokens(
      await fixture.alca.getAddress(),
      await fixture.publicStaking.getAddress()
    );
    await fixture.alca.transfer(
      await reentrantLoopCollectEthERC721ReceiverAccount.getAddress(),
      1000
    );
    await reentrantLoopCollectEthERC721ReceiverAccount.approve(
      await fixture.publicStaking.getAddress(),
      1000
    );

    // mint to another user with the same amount
    await fixture.alca.approve(await fixture.publicStaking.getAddress(), 1000);
    const [adminSigner] = await ethers.getSigners();
    await fixture.publicStaking.mintTo(adminSigner.address, 1000, 1);

    // mint a position with the reentrancy user and mine 2 blocks
    await reentrantLoopCollectEthERC721ReceiverAccount.mint(1000);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.parseEther("1000"),
    });
    const tokenID = 1;

    expect(
      await ethers.provider.getBalance(
        await reentrantLoopCollectEthERC721ReceiverAccount.getAddress()
      )
    ).to.equals(
      ethers.parseEther("0"),
      "Eth amount not matched expected amount!"
    );

    expect(
      (await fixture.publicStaking.ownerOf(tokenID)).toLowerCase()
    ).to.be.equals(adminSigner.address.toLowerCase());
    expect(
      await fixture.publicStaking.balanceOf(adminSigner.address)
    ).to.be.equals(1n);

    expect(
      await fixture.publicStaking.balanceOf(
        await reentrantLoopCollectEthERC721ReceiverAccount.getAddress()
      )
    ).to.be.equals(1n);

    await fixture.publicStaking
      .connect(adminSigner)
      ["safeTransferFrom(address,address,uint256)"](
        adminSigner.address,
        await reentrantLoopCollectEthERC721ReceiverAccount.getAddress(),
        tokenID
      );

    expect(
      (await fixture.publicStaking.ownerOf(tokenID)).toLowerCase()
    ).to.be.equals(
      (
        await reentrantLoopCollectEthERC721ReceiverAccount.getAddress()
      ).toLowerCase()
    );
    expect(
      await fixture.publicStaking.balanceOf(adminSigner.address)
    ).to.be.equals(0n);
    expect(
      await fixture.publicStaking.balanceOf(
        await reentrantLoopCollectEthERC721ReceiverAccount.getAddress()
      )
    ).to.be.equals(2n);

    // the user should get the expected amount not a cent more
    expect(
      await ethers.provider.getBalance(
        await reentrantLoopCollectEthERC721ReceiverAccount.getAddress()
      )
    ).to.equals(
      ethers.parseEther("500"),
      "Eth amount not matched expected amount!"
    );
  });
});
