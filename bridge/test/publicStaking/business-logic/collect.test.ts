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
  collectTokensCheckAndUpdateState,
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
    const baseAmount = ethers.utils.parseUnits("100", 0).toBigInt();
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

  it("Mint, collect and burn tokens for multiple users", async function () {
    const sharesPerUser = 100n;
    const tokensID: number[] = [];
    for (let i = 0; i < users.length; i++) {
      tokensID.push(0);
    }

    const scaleFactor = (
      await fixture.publicStaking.getAccumulatorScaleFactor()
    ).toBigInt();

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

    // deposit and collect the profits for all users
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

    // mint another position.
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

    // All users collect this time, we expect the last user witch didn't register last time to get more
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

    await burnPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.madToken,
      333_333333333333333333n,
      100n,
      0,
      users,
      tokensID,
      expectedState,
      "After burn 1"
    );
  });
});
