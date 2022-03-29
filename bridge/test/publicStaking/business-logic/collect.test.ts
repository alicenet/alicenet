import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import {
  BaseTokensFixture,
  createUsers,
  getBaseTokensFixture,
  getTokenIdFromTx,
  mineBlocks,
} from "../../setup";
import {
  burnPositionCheckAndUpdateState,
  collectEthCheckAndUpdateState,
  collectEthToCheckAndUpdateState,
  collectTokensCheckAndUpdateState,
  collectTokensToCheckAndUpdateState,
  depositEthCheckAndUpdateState,
  depositTokensCheckAndUpdateState,
  getCurrentState,
  mintPositionCheckAndUpdateState,
} from "../setup";

describe("PublicStaking: Collect Tokens and ETH profit", async () => {
  let fixture: BaseTokensFixture;
  let users: SignerWithAddress[];
  const numberUsers = 3;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    await fixture.madToken.approve(
      fixture.publicStaking.address,
      ethers.utils.parseUnits("100000", 18)
    );
    users = await createUsers(numberUsers);
    const baseAmount = ethers.utils.parseUnits("100", 18).toBigInt();
    for (let i = 0; i < numberUsers; i++) {
      await fixture.madToken.transfer(await users[i].getAddress(), baseAmount);
      await fixture.madToken
        .connect(users[i])
        .approve(fixture.publicStaking.address, baseAmount);
    }
    await mineBlocks(2n);
  });

  it("Shouldn't allow to collect funds before time", async function () {
    const sharesPerUser = 100n;
    const tx = await fixture.publicStaking
      .connect(users[0])
      .mint(sharesPerUser);
    const tokenID = await getTokenIdFromTx(tx);
    await expect(
      fixture.publicStaking.connect(users[0]).collectEth(tokenID)
    ).to.revertedWith("PublicStaking: Cannot withdraw at the moment.");
  });

  it("Shouldn't allow to collect funds for not owned position", async function () {
    const sharesPerUser = 100n;
    const tx = await fixture.publicStaking
      .connect(users[0])
      .mint(sharesPerUser);
    const tokenID = await getTokenIdFromTx(tx);
    await expect(fixture.publicStaking.collectEth(tokenID)).to.revertedWith(
      "PublicStaking: Error sender is not the owner of the tokenID!"
    );
    await expect(fixture.publicStaking.collectToken(tokenID)).to.revertedWith(
      "PublicStaking: Error sender is not the owner of the tokenID!"
    );
  });

  it("Anyone can estimate profits for a position", async function () {
    const sharesPerUser = 100n;
    const tx = await fixture.publicStaking
      .connect(users[0])
      .mint(sharesPerUser);
    const tokenID = await getTokenIdFromTx(tx);
    // Non user tries to estimate the profit of a non owned position
    await fixture.publicStaking.estimateEthCollection(tokenID);
    await fixture.publicStaking.estimateTokenCollection(tokenID);
  });

  it("Shouldn't allow to collect funds for non-existing position", async function () {
    await expect(fixture.publicStaking.collectEth(100)).to.revertedWith(
      "ERC721: owner query for nonexistent token"
    );
    await expect(fixture.publicStaking.collectToken(100)).to.revertedWith(
      "ERC721: owner query for nonexistent token"
    );
  });

  it("Shouldn't allow to estimate funds for non-existing position", async function () {
    await expect(
      fixture.publicStaking.estimateEthCollection(100)
    ).to.revertedWith("PublicStaking: Error, NFT token doesn't exist!");
    await expect(
      fixture.publicStaking.estimateTokenCollection(100)
    ).to.revertedWith("PublicStaking: Error, NFT token doesn't exist!");
  });

  it("Mint and collect profits", async function () {
    const sharesPerUser = ethers.utils.parseUnits("100", 18).toBigInt();
    const tokensID: number[] = [];
    for (let i = 0; i < users.length; i++) {
      tokensID.push(0);
    }

    const expectedState = await getCurrentState(
      fixture.publicStaking,
      fixture.madToken,
      users,
      tokensID
    );

    await mintPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      sharesPerUser,
      0,
      users,
      tokensID,
      expectedState,
      "After mint 1"
    );

    await mintPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      sharesPerUser,
      1,
      users,
      tokensID,
      expectedState,
      "After mint 1"
    );

    // deposit and collect only with 1 user
    const amountDeposited = ethers.utils.parseUnits("50", 18).toBigInt();
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1"
    );

    await depositEthCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 Eth"
    );

    for (let i = 0; i < 2; i++) {
      await collectTokensCheckAndUpdateState(
        fixture.publicStaking,
        fixture.madToken,
        amountDeposited / 2n,
        i,
        users,
        tokensID,
        expectedState,
        "After collect 1" + i
      );

      await collectEthCheckAndUpdateState(
        fixture.publicStaking,
        fixture.madToken,
        amountDeposited / 2n,
        i,
        users,
        tokensID,
        expectedState,
        "After collect 1 Eth" + i
      );
    }
  });

  it("Mint and collectTo profits", async function () {
    const sharesPerUser = ethers.utils.parseUnits("100", 18).toBigInt();
    const tokensID: number[] = [];
    for (let i = 0; i < users.length; i++) {
      tokensID.push(0);
    }

    const expectedState = await getCurrentState(
      fixture.publicStaking,
      fixture.madToken,
      users,
      tokensID
    );

    await mintPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      sharesPerUser,
      0,
      users,
      tokensID,
      expectedState,
      "After mint 1"
    );

    // deposit and collect only with 1 user
    const amountDeposited = ethers.utils.parseUnits("50", 18).toBigInt();
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1"
    );

    await depositEthCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 Eth"
    );

    const balanceBeforeDestination = (
      await ethers.provider.getBalance(users[1].address)
    ).toBigInt();
    await collectTokensToCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      0,
      1,
      users,
      tokensID,
      expectedState,
      "After collectTo 1"
    );

    await collectEthToCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      0,
      1,
      users,
      tokensID,
      expectedState,
      "After collectTo 1 Eth"
    );

    // only the destination user has to have received the profit
    expect(
      (await ethers.provider.getBalance(users[1].address)).toBigInt()
    ).to.be.equals(
      balanceBeforeDestination + amountDeposited,
      "Expected ETH not met for destination address"
    );
  });

  it("Mint, collect and burn tokens for 3 users", async function () {
    const sharesPerUser = 100n;
    const tokensID: number[] = [];
    for (let i = 0; i < users.length; i++) {
      tokensID.push(0);
    }

    const expectedState = await getCurrentState(
      fixture.publicStaking,
      fixture.madToken,
      users,
      tokensID
    );

    await mintPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      sharesPerUser,
      0,
      users,
      tokensID,
      expectedState,
      "After mint 1"
    );

    // deposit and collect only with 1 user
    let amountDeposited = 50n;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1"
    );

    // user 1 should get all the amount deposited since he's the only that skated in the contract
    await collectTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      0,
      users,
      tokensID,
      expectedState,
      "After collect 1"
    );

    // mint another position.
    await mintPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      sharesPerUser,
      1,
      users,
      tokensID,
      expectedState,
      "After mint 2"
    );

    // deposit and collect the profits for all users. Since each user staked a equal amount, each user
    // should receive 50% of the amount deposited
    amountDeposited = 500n;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2"
    );

    let numUsers = 2;
    for (let i = 0; i < numUsers; i++) {
      // 50% of the amount deposited
      const expectedCollectedAmount = amountDeposited / BigInt(numUsers);
      await collectTokensCheckAndUpdateState(
        fixture.publicStaking,
        fixture.madToken,
        expectedCollectedAmount,
        i,
        users,
        tokensID,
        expectedState,
        "After collect 2-" + i
      );
    }

    // mint another position. With 3 users that staked the same amount, each user should receive 33.33%
    // of the amount deposited, and we should see some leftovers in the slush
    await mintPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      sharesPerUser,
      2,
      users,
      tokensID,
      expectedState,
      "After mint 3"
    );

    // deposit and collect the profits for all users
    amountDeposited = 1000n;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 3"
    );

    let expectedSlushes = [
      333333333333333400n,
      666666666666666700n,
      1000000000000000000n,
    ];
    numUsers = 3;
    for (let i = 0; i < numUsers; i++) {
      const expectedCollectedAmount = amountDeposited / BigInt(numUsers);
      await collectTokensCheckAndUpdateState(
        fixture.publicStaking,
        fixture.madToken,
        expectedCollectedAmount,
        i,
        users,
        tokensID,
        expectedState,
        "After collect 3-" + i,
        expectedSlushes[i]
      );
    }

    // deposit and collect the profits for all users
    amountDeposited = 1000n;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 4"
    );

    // Checking the logic if only 2 users withdrawal. The 3rd user will withdrawal later, and should
    // receive the cumulative amount of the 2 deposits without any lost
    expectedSlushes = [666666666666666800n, 1333333333333333400n];

    // only 2 users collect
    for (let i = 0; i < 2; i++) {
      const expectedCollectedAmount = amountDeposited / BigInt(numUsers);
      await collectTokensCheckAndUpdateState(
        fixture.publicStaking,
        fixture.madToken,
        expectedCollectedAmount,
        i,
        users,
        tokensID,
        expectedState,
        "After collect 4-" + i,
        expectedSlushes[i]
      );
    }

    // deposit and collect the profits for all users
    amountDeposited = 1666n;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 5"
    );

    // All users collect this time, we expect the last user witch didn't withdrawal last time to get more
    const expectedCollectedAmounts = [555n, 555n, 889n];
    expectedSlushes = [
      777777777777777800n,
      1555555555555555600n,
      2000000000000000000n,
    ];
    for (let i = 0; i < numUsers; i++) {
      await collectTokensCheckAndUpdateState(
        fixture.publicStaking,
        fixture.madToken,
        expectedCollectedAmounts[i],
        i,
        users,
        tokensID,
        expectedState,
        "After collect 4-" + i,
        expectedSlushes[i]
      );
    }
    await mineBlocks(2n);

    // deposit eth this time
    amountDeposited = ethers.utils.parseEther("1000").toBigInt();
    await depositEthCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 6 Eth"
    );
    let expectedSlushEth = 333333333333333400n;

    // Start to burn the positions
    await burnPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      sharesPerUser,
      333_333333333333333333n,
      100n,
      0,
      users,
      tokensID,
      expectedState,
      "After burn 1",
      expectedSlushEth
    );

    expectedSlushEth = 666666666666666700n;
    await burnPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      sharesPerUser,
      333_333333333333333333n,
      100n,
      1,
      users,
      tokensID,
      expectedState,
      "After burn 2",
      expectedSlushEth
    );

    // last user should get the shares + slushes since he's the last user exiting the staking contract
    expectedSlushEth = 0n;
    const expectedSlushToken = 0n;
    await burnPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      sharesPerUser,
      333_333333333333333334n,
      102n,
      2,
      users,
      tokensID,
      expectedState,
      "After burn 3",
      expectedSlushEth,
      expectedSlushToken
    );
  });
});
