import { ethers } from "hardhat";
import {
  BaseTokensFixture,
  expect,
  getBaseTokensFixture,
  mineBlocks,
} from "../../setup";

describe("PublicStaking: Reentrancy tests", async () => {
  let fixture: BaseTokensFixture;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
  });

  it("Should not allow reentrancy in the CollectEth", async function () {
    const ReentrantLoopEthCollectorAccount = await ethers.getContractFactory(
      "ReentrantLoopEthCollectorAccount"
    );
    const reentrantLoopEthCollectorAccount =
      await ReentrantLoopEthCollectorAccount.deploy();
    await reentrantLoopEthCollectorAccount.setTokens(
      fixture.madToken.address,
      fixture.publicStaking.address
    );
    await fixture.madToken.transfer(
      reentrantLoopEthCollectorAccount.address,
      1000
    );
    await reentrantLoopEthCollectorAccount.approve(
      fixture.publicStaking.address,
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks
    await reentrantLoopEthCollectorAccount.mint(1000);

    // mint to another user with the same amount
    await fixture.madToken.approve(fixture.publicStaking.address, 1000);
    await fixture.publicStaking.mint(1000);
    await mineBlocks(2n);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.utils.parseEther("1000"),
    });
    const tokenID = 1;
    const balanceBefore = (
      await ethers.provider.getBalance(reentrantLoopEthCollectorAccount.address)
    ).toBigInt();
    await reentrantLoopEthCollectorAccount.setTokenID(tokenID);
    // The collectEth reentrancy should only be called once, since the second time all the profit was
    // already transferred and from _safeTransferEth function we just return when the amount is 0
    await reentrantLoopEthCollectorAccount.collectEth(tokenID);
    // the balance of the dishonest user should only have increased by 500 ether (the right amount). If
    // this is not true it means that there was an exploit!
    expect(
      (
        await ethers.provider.getBalance(
          reentrantLoopEthCollectorAccount.address
        )
      ).toBigInt()
    ).to.be.equals(
      balanceBefore + ethers.utils.parseEther("500").toBigInt(),
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
      fixture.madToken.address,
      fixture.publicStaking.address
    );
    await fixture.madToken.transfer(
      reentrantFiniteEthCollectorAccount.address,
      1000
    );
    await reentrantFiniteEthCollectorAccount.approve(
      fixture.publicStaking.address,
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks
    await reentrantFiniteEthCollectorAccount.mint(1000);

    // mint to another user with the same amount
    await fixture.madToken.approve(fixture.publicStaking.address, 1000);
    await fixture.publicStaking.mint(1000);
    await mineBlocks(2n);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.utils.parseEther("1000"),
    });
    const tokenID = 1;
    const balanceBefore = (
      await ethers.provider.getBalance(
        reentrantFiniteEthCollectorAccount.address
      )
    ).toBigInt();
    await reentrantFiniteEthCollectorAccount.setTokenID(tokenID);
    // The collectEth reentrancy should only be called once, since the second time all the profit was
    // already transferred and from _safeTransferEth function we just return when the amount is 0
    await reentrantFiniteEthCollectorAccount.collectEth(tokenID);
    // the balance of the dishonest user should only have increased by 500 ether (the right amount). If
    // this is not true it means that there was an exploit!
    expect(
      (
        await ethers.provider.getBalance(
          reentrantFiniteEthCollectorAccount.address
        )
      ).toBigInt()
    ).to.be.equals(
      balanceBefore + ethers.utils.parseEther("500").toBigInt(),
      "Balance after reentracy doesn't match!"
    );
  });

  it("Should not allow reentrancy when burning a position", async function () {
    const ReentrantLoopBurnAccount = await ethers.getContractFactory(
      "ReentrantLoopBurnAccount"
    );
    const reentrantLoopBurnAccount = await ReentrantLoopBurnAccount.deploy();
    await reentrantLoopBurnAccount.setTokens(
      fixture.madToken.address,
      fixture.publicStaking.address
    );
    await fixture.madToken.transfer(reentrantLoopBurnAccount.address, 1000);
    await reentrantLoopBurnAccount.approve(fixture.publicStaking.address, 1000);
    // mint a position with the reentrancy user and mine 2 blocks
    await reentrantLoopBurnAccount.mint(1000);

    // mint to another user with the same amount
    await fixture.madToken.approve(fixture.publicStaking.address, 1000);
    await fixture.publicStaking.mint(1000);
    await mineBlocks(2n);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.utils.parseEther("1000"),
    });
    const tokenID = 1;
    await reentrantLoopBurnAccount.setTokenID(tokenID);

    await expect(reentrantLoopBurnAccount.burn(tokenID)).to.be.rejectedWith(
      "EthSafeTransfer: Transfer failed."
    );
  });

  it("Should not allow reentrancy when burning a position with finite loop", async function () {
    const ReentrantFiniteBurnAccount = await ethers.getContractFactory(
      "ReentrantFiniteBurnAccount"
    );
    const reentrantFiniteBurnAccount =
      await ReentrantFiniteBurnAccount.deploy();
    await reentrantFiniteBurnAccount.setTokens(
      fixture.madToken.address,
      fixture.publicStaking.address
    );
    await fixture.madToken.transfer(reentrantFiniteBurnAccount.address, 1000);
    await reentrantFiniteBurnAccount.approve(
      fixture.publicStaking.address,
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks
    await reentrantFiniteBurnAccount.mint(1000);

    // mint to another user with the same amount
    await fixture.madToken.approve(fixture.publicStaking.address, 1000);
    await fixture.publicStaking.mint(1000);
    await mineBlocks(2n);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.utils.parseEther("1000"),
    });
    const tokenID = 1;
    await reentrantFiniteBurnAccount.setTokenID(tokenID);

    await expect(reentrantFiniteBurnAccount.burn(tokenID)).to.be.rejectedWith(
      "EthSafeTransfer: Transfer failed."
    );
  });

  it("Should not allow burn reentrancy after transferring a NFT position", async function () {
    const ReentrantLoopBurnERC721ReceiverAccount =
      await ethers.getContractFactory("ReentrantLoopBurnERC721ReceiverAccount");
    const reentrantLoopBurnERC721ReceiverAccount =
      await ReentrantLoopBurnERC721ReceiverAccount.deploy();
    await reentrantLoopBurnERC721ReceiverAccount.setTokens(
      fixture.madToken.address,
      fixture.publicStaking.address
    );
    await fixture.madToken.transfer(
      reentrantLoopBurnERC721ReceiverAccount.address,
      1000
    );
    await reentrantLoopBurnERC721ReceiverAccount.approve(
      fixture.publicStaking.address,
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks

    // mint to another user with the same amount
    await fixture.madToken.approve(fixture.publicStaking.address, 1000);
    await fixture.publicStaking.mint(1000);
    await reentrantLoopBurnERC721ReceiverAccount.mint(1000);
    await mineBlocks(2n);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.utils.parseEther("1000"),
    });
    const tokenID = 1;

    const [adminSigner] = await ethers.getSigners();

    await expect(
      fixture.publicStaking["safeTransferFrom(address,address,uint256)"](
        adminSigner.address,
        reentrantLoopBurnERC721ReceiverAccount.address,
        tokenID
      )
    ).to.be.rejectedWith("EthSafeTransfer: Transfer failed.");
  });

  it("Should not allow burn reentrancy after transferring a NFT position finite loop", async function () {
    const ReentrantFiniteBurnERC721ReceiverAccount =
      await ethers.getContractFactory(
        "ReentrantFiniteBurnERC721ReceiverAccount"
      );
    const reentrantFiniteBurnERC721ReceiverAccount =
      await ReentrantFiniteBurnERC721ReceiverAccount.deploy();
    await reentrantFiniteBurnERC721ReceiverAccount.setTokens(
      fixture.madToken.address,
      fixture.publicStaking.address
    );
    await fixture.madToken.transfer(
      reentrantFiniteBurnERC721ReceiverAccount.address,
      1000
    );
    await reentrantFiniteBurnERC721ReceiverAccount.approve(
      fixture.publicStaking.address,
      1000
    );
    // mint a position with the reentrancy user and mine 2 blocks

    // mint to another user with the same amount
    await fixture.madToken.approve(fixture.publicStaking.address, 1000);
    await fixture.publicStaking.mint(1000);
    await reentrantFiniteBurnERC721ReceiverAccount.mint(1000);
    await mineBlocks(2n);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.utils.parseEther("1000"),
    });
    const tokenID = 1;

    const [adminSigner] = await ethers.getSigners();

    await expect(
      fixture.publicStaking["safeTransferFrom(address,address,uint256)"](
        adminSigner.address,
        reentrantFiniteBurnERC721ReceiverAccount.address,
        tokenID
      )
    ).to.be.rejectedWith("EthSafeTransfer: Transfer failed.");
  });

  it("Should not allow collect reentrancy after transferring a NFT position", async function () {
    const ReentrantLoopCollectEthERC721ReceiverAccount =
      await ethers.getContractFactory(
        "ReentrantLoopCollectEthERC721ReceiverAccount"
      );
    const reentrantLoopCollectEthERC721ReceiverAccount =
      await ReentrantLoopCollectEthERC721ReceiverAccount.deploy();
    await reentrantLoopCollectEthERC721ReceiverAccount.setTokens(
      fixture.madToken.address,
      fixture.publicStaking.address
    );
    await fixture.madToken.transfer(
      reentrantLoopCollectEthERC721ReceiverAccount.address,
      1000
    );
    await reentrantLoopCollectEthERC721ReceiverAccount.approve(
      fixture.publicStaking.address,
      1000
    );

    // mint to another user with the same amount
    await fixture.madToken.approve(fixture.publicStaking.address, 1000);
    const [adminSigner] = await ethers.getSigners();
    await fixture.publicStaking.mintTo(adminSigner.address, 1000, 1);

    // mint a position with the reentrancy user and mine 2 blocks
    await reentrantLoopCollectEthERC721ReceiverAccount.mint(1000);
    await mineBlocks(2n);

    // someone deposits eth
    await fixture.publicStaking.depositEth(42, {
      value: ethers.utils.parseEther("1000"),
    });
    const tokenID = 1;

    expect(
      (
        await ethers.provider.getBalance(
          reentrantLoopCollectEthERC721ReceiverAccount.address
        )
      ).toBigInt()
    ).to.equals(
      ethers.utils.parseEther("0").toBigInt(),
      "Eth amount not matched expected amount!"
    );

    expect(
      (await fixture.publicStaking.ownerOf(tokenID)).toLowerCase()
    ).to.be.equals(adminSigner.address.toLowerCase());
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(1n);

    expect(
      (
        await fixture.publicStaking.balanceOf(
          reentrantLoopCollectEthERC721ReceiverAccount.address
        )
      ).toBigInt()
    ).to.be.equals(1n);

    await fixture.publicStaking
      .connect(adminSigner)
      ["safeTransferFrom(address,address,uint256)"](
        adminSigner.address,
        reentrantLoopCollectEthERC721ReceiverAccount.address,
        tokenID
      );

    expect(
      (await fixture.publicStaking.ownerOf(tokenID)).toLowerCase()
    ).to.be.equals(
      reentrantLoopCollectEthERC721ReceiverAccount.address.toLowerCase()
    );
    expect(
      (await fixture.publicStaking.balanceOf(adminSigner.address)).toBigInt()
    ).to.be.equals(0n);
    expect(
      (
        await fixture.publicStaking.balanceOf(
          reentrantLoopCollectEthERC721ReceiverAccount.address
        )
      ).toBigInt()
    ).to.be.equals(2n);

    // the user should get the expected amount not a cent more
    expect(
      (
        await ethers.provider.getBalance(
          reentrantLoopCollectEthERC721ReceiverAccount.address
        )
      ).toBigInt()
    ).to.equals(
      ethers.utils.parseEther("500").toBigInt(),
      "Eth amount not matched expected amount!"
    );
  });
});
