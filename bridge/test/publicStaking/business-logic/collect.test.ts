import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import {
  BaseTokensFixture,
  callFunctionAndGetReturnValues,
  createUsers,
  getBaseTokensFixture,
  getTokenIdFromTx,
  mineBlocks,
} from "../../setup";
import {
  burnPositionCheckAndUpdateState,
  collectAllProfitsCheckAndUpdateState,
  collectAllProfitsToCheckAndUpdateState,
  collectEthCheckAndUpdateState,
  collectEthToCheckAndUpdateState,
  collectTokensCheckAndUpdateState,
  collectTokensToCheckAndUpdateState,
  depositEthCheckAndUpdateState,
  depositTokensCheckAndUpdateState,
  getCurrentState,
  mintPositionCheckAndUpdateState,
  StakingState,
} from "../setup";

describe("PublicStaking: Collect Tokens and ETH profit", async () => {
  let fixture: BaseTokensFixture;
  let users: SignerWithAddress[];
  const numberUsers = 3;
  let admin: SignerWithAddress;
  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    [admin] = await ethers.getSigners();
    await fixture.aToken.approve(
      fixture.publicStaking.address,
      ethers.utils.parseUnits("100000", 18)
    );
    users = await createUsers(numberUsers);
    const baseAmount = ethers.utils.parseUnits("100", 18).toBigInt();
    for (let i = 0; i < numberUsers; i++) {
      await fixture.aToken.transfer(await users[i].getAddress(), baseAmount);
      await fixture.aToken
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
    ).to.be.revertedWithCustomError(
      fixture.publicStaking,
      "LockDurationWithdrawTimeNotReached"
    );
  });

  it("Shouldn't allow to collect funds for not owned position", async function () {
    const sharesPerUser = 100n;
    const tx = await fixture.publicStaking
      .connect(users[0])
      .mint(sharesPerUser);
    const tokenID = await getTokenIdFromTx(tx);
    await expect(fixture.publicStaking.collectEth(tokenID))
      .to.be.revertedWithCustomError(
        fixture.publicStaking,
        "CallerNotTokenOwner"
      )
      .withArgs(admin.address);
    await expect(fixture.publicStaking.collectToken(tokenID))
      .to.be.revertedWithCustomError(
        fixture.publicStaking,
        "CallerNotTokenOwner"
      )
      .withArgs(admin.address);
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
      "ERC721: invalid token ID"
    );
    await expect(fixture.publicStaking.collectToken(100)).to.revertedWith(
      "ERC721: invalid token ID"
    );
  });

  it("Shouldn't allow to estimate funds for non-existing position", async function () {
    const nonExistingTokenId = 100;
    await expect(
      fixture.publicStaking.estimateEthCollection(nonExistingTokenId)
    )
      .to.be.revertedWithCustomError(fixture.publicStaking, "InvalidTokenId")
      .withArgs(nonExistingTokenId);
    await expect(fixture.publicStaking.estimateTokenCollection(100))
      .to.be.revertedWithCustomError(fixture.publicStaking, "InvalidTokenId")
      .withArgs(nonExistingTokenId);
  });

  describe("with positions minted", async () => {
    let sharesPerUser: bigint;
    let expectedState: StakingState;
    let tokensID: number[];
    beforeEach(async function () {
      sharesPerUser = ethers.utils.parseUnits("100", 18).toBigInt();

      tokensID = [];
      for (let i = 0; i < users.length; i++) {
        tokensID.push(0);
      }

      expectedState = await getCurrentState(
        fixture.publicStaking,
        fixture.aToken,
        users,
        tokensID
      );
    });

    it("collect profits individually", async function () {
      await mintPositionCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        sharesPerUser,
        0,
        users,
        tokensID,
        expectedState,
        "After mint 1"
      );

      await mintPositionCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
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
        fixture.aToken,
        amountDeposited,
        users,
        tokensID,
        expectedState,
        "After deposit 1"
      );

      await depositEthCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        amountDeposited,
        users,
        tokensID,
        expectedState,
        "After deposit 1 Eth"
      );

      for (let i = 0; i < 2; i++) {
        await collectTokensCheckAndUpdateState(
          fixture.publicStaking,
          fixture.aToken,
          amountDeposited / 2n,
          i,
          users,
          tokensID,
          expectedState,
          "After collect 1" + i
        );

        await collectEthCheckAndUpdateState(
          fixture.publicStaking,
          fixture.aToken,
          amountDeposited / 2n,
          i,
          users,
          tokensID,
          expectedState,
          "After collect 1 Eth" + i
        );
      }
    });

    it("collect all profits", async function () {
      await mintPositionCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        sharesPerUser,
        0,
        users,
        tokensID,
        expectedState,
        "After mint 1"
      );

      await mintPositionCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
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
        fixture.aToken,
        amountDeposited,
        users,
        tokensID,
        expectedState,
        "After deposit 1"
      );

      await depositEthCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        amountDeposited,
        users,
        tokensID,
        expectedState,
        "After deposit 1 Eth"
      );

      for (let i = 0; i < 2; i++) {
        await collectAllProfitsCheckAndUpdateState(
          fixture.publicStaking,
          fixture.aToken,
          amountDeposited / 2n,
          amountDeposited / 2n,
          i,
          users,
          tokensID,
          expectedState,
          "After collect all profits " + i
        );
      }
    });

    it("collectTo profits individually", async function () {
      await mintPositionCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
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
        fixture.aToken,
        amountDeposited,
        users,
        tokensID,
        expectedState,
        "After deposit 1"
      );

      await depositEthCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
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
        fixture.aToken,
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
        fixture.aToken,
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

    it("collect all profits at once to", async function () {
      await mintPositionCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
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
        fixture.aToken,
        amountDeposited,
        users,
        tokensID,
        expectedState,
        "After deposit 1"
      );

      await depositEthCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        amountDeposited,
        users,
        tokensID,
        expectedState,
        "After deposit 1 Eth"
      );

      const balanceBeforeDestination = (
        await ethers.provider.getBalance(users[1].address)
      ).toBigInt();

      await collectAllProfitsToCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        amountDeposited,
        amountDeposited,
        0,
        1,
        users,
        tokensID,
        expectedState,
        "After collectAllProfitsTo 1"
      );

      // only the destination user has to have received the profit
      expect(
        (await ethers.provider.getBalance(users[1].address)).toBigInt()
      ).to.be.equals(
        balanceBeforeDestination + amountDeposited,
        "Expected ETH not met for destination address"
      );
    });
  });

  it("Mint, collect and burn tokens for 3 users", async function () {
    const scaleFactor = (
      await fixture.publicStaking.getAccumulatorScaleFactor()
    ).toBigInt();
    const sharesPerUser = 100n;
    const tokensID: number[] = [];
    for (let i = 0; i < users.length; i++) {
      tokensID.push(0);
    }
    let tokenStateAccum = 0n;
    let tokenStateSlush = 0n;

    const expectedState = await getCurrentState(
      fixture.publicStaking,
      fixture.aToken,
      users,
      tokensID
    );

    await mintPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      sharesPerUser,
      0,
      users,
      tokensID,
      expectedState,
      "After mint 1"
    );

    // deposit and collect only with 1 user
    let amountDeposited = 50n;
    tokenStateSlush += 50n * scaleFactor;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1"
    );

    // Perform slushSkim
    let totalSharesBN = await fixture.publicStaking.getTotalShares();
    let totalShares = totalSharesBN.toBigInt();
    let deltaAccum = tokenStateSlush / totalShares;
    tokenStateSlush -= deltaAccum * totalShares;
    tokenStateAccum += deltaAccum;
    // get position info
    let [userPosition] = await callFunctionAndGetReturnValues(
      fixture.publicStaking,
      "getPosition",
      users[0] as SignerWithAddress,
      [tokensID[0]]
    );
    // compute payout
    let diffAccum = tokenStateAccum - userPosition.accumulatorToken.toBigInt();
    let payoutEst = diffAccum * sharesPerUser;
    if (totalShares == userPosition.shares.toBigInt()) {
      payoutEst += tokenStateSlush;
      tokenStateSlush = 0n;
    }
    let payoutRem = payoutEst;
    payoutEst /= scaleFactor;
    payoutRem -= payoutEst * scaleFactor;
    tokenStateSlush += payoutRem;

    // user 1 should get all the amount deposited since he's the only that staked in the contract
    await collectTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      payoutEst,
      0,
      users,
      tokensID,
      expectedState,
      "After collect 1"
    );

    // Perform slushSkim
    totalSharesBN = await fixture.publicStaking.getTotalShares();
    totalShares = totalSharesBN.toBigInt();
    deltaAccum = tokenStateSlush / totalShares;
    tokenStateSlush -= deltaAccum * totalShares;
    tokenStateAccum += deltaAccum;

    // mint another position.
    await mintPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
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
    tokenStateSlush += 500n * scaleFactor;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2"
    );

    let numUsers = 2;
    for (let i = 0; i < numUsers; i++) {
      // Perform slushSkim
      let totalSharesBN = await fixture.publicStaking.getTotalShares();
      let totalShares = totalSharesBN.toBigInt();
      let deltaAccum = tokenStateSlush / totalShares;
      tokenStateSlush -= deltaAccum * totalShares;
      tokenStateAccum += deltaAccum;
      // get position info
      const [userPosition] = await callFunctionAndGetReturnValues(
        fixture.publicStaking,
        "getPosition",
        users[i] as SignerWithAddress,
        [tokensID[i]]
      );
      // compute payout
      let diffAccum =
        tokenStateAccum - userPosition.accumulatorToken.toBigInt();
      let payoutEst = diffAccum * sharesPerUser;
      if (totalShares == userPosition.shares.toBigInt()) {
        payoutEst += tokenStateSlush;
        tokenStateSlush = 0n;
      }
      let payoutRem = payoutEst;
      payoutEst /= scaleFactor;
      payoutRem -= payoutEst * scaleFactor;
      tokenStateSlush += payoutRem;

      // 50% of the amount deposited
      await collectTokensCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        payoutEst,
        i,
        users,
        tokensID,
        expectedState,
        "After collect 2-" + i
      );
    }

    // Perform slushSkim
    totalSharesBN = await fixture.publicStaking.getTotalShares();
    totalShares = totalSharesBN.toBigInt();
    deltaAccum = tokenStateSlush / totalShares;
    tokenStateSlush -= deltaAccum * totalShares;
    tokenStateAccum += deltaAccum;

    // mint another position. With 3 users that staked the same amount, each user should receive 33.33%
    // of the amount deposited, and we should see some leftovers in the slush
    await mintPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      sharesPerUser,
      2,
      users,
      tokensID,
      expectedState,
      "After mint 3"
    );

    // deposit and collect the profits for all users
    amountDeposited = 1000n;
    tokenStateSlush += 1000n * scaleFactor;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 3"
    );

    numUsers = 3;
    for (let i = 0; i < numUsers; i++) {
      // Perform slushSkim
      let totalSharesBN = await fixture.publicStaking.getTotalShares();
      let totalShares = totalSharesBN.toBigInt();
      let deltaAccum = tokenStateSlush / totalShares;
      tokenStateSlush -= deltaAccum * totalShares;
      tokenStateAccum += deltaAccum;
      // get position info
      const [userPosition] = await callFunctionAndGetReturnValues(
        fixture.publicStaking,
        "getPosition",
        users[i] as SignerWithAddress,
        [tokensID[i]]
      );
      // compute payout
      let diffAccum =
        tokenStateAccum - userPosition.accumulatorToken.toBigInt();
      let payoutEst = diffAccum * sharesPerUser;
      if (totalShares == userPosition.shares.toBigInt()) {
        payoutEst += tokenStateSlush;
        tokenStateSlush = 0n;
      }
      let payoutRem = payoutEst;
      payoutEst /= scaleFactor;
      payoutRem -= payoutEst * scaleFactor;
      tokenStateSlush += payoutRem;

      await collectTokensCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        payoutEst,
        i,
        users,
        tokensID,
        expectedState,
        "After collect 3-" + i,
        tokenStateSlush
      );
    }

    // deposit and collect the profits for all users
    amountDeposited = 1000n;
    tokenStateSlush += 1000n * scaleFactor;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 4"
    );

    // Checking the logic if only 2 users withdrawal. The 3rd user will withdrawal later, and should
    // receive the cumulative amount of the 2 deposits without any lost

    // only 2 users collect
    for (let i = 0; i < 2; i++) {
      // Perform slushSkim
      let totalSharesBN = await fixture.publicStaking.getTotalShares();
      let totalShares = totalSharesBN.toBigInt();
      let deltaAccum = tokenStateSlush / totalShares;
      tokenStateSlush -= deltaAccum * totalShares;
      tokenStateAccum += deltaAccum;
      // get position info
      const [userPosition] = await callFunctionAndGetReturnValues(
        fixture.publicStaking,
        "getPosition",
        users[i] as SignerWithAddress,
        [tokensID[i]]
      );
      // compute payout
      let diffAccum =
        tokenStateAccum - userPosition.accumulatorToken.toBigInt();
      let payoutEst = diffAccum * sharesPerUser;
      if (totalShares == userPosition.shares.toBigInt()) {
        payoutEst += tokenStateSlush;
        tokenStateSlush = 0n;
      }
      let payoutRem = payoutEst;
      payoutEst /= scaleFactor;
      payoutRem -= payoutEst * scaleFactor;
      tokenStateSlush += payoutRem;

      await collectTokensCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        payoutEst,
        i,
        users,
        tokensID,
        expectedState,
        "After collect 4-" + i,
        tokenStateSlush
      );
    }

    // deposit and collect the profits for all users
    amountDeposited = 1666n;
    tokenStateSlush += 1666n * scaleFactor;
    await depositTokensCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 5"
    );

    // All users collect this time, we expect the last user witch didn't withdrawal last time to get more
    for (let i = 0; i < numUsers; i++) {
      // Perform slushSkim
      let totalSharesBN = await fixture.publicStaking.getTotalShares();
      let totalShares = totalSharesBN.toBigInt();
      let deltaAccum = tokenStateSlush / totalShares;
      tokenStateSlush -= deltaAccum * totalShares;
      tokenStateAccum += deltaAccum;
      // get position info
      const [userPosition] = await callFunctionAndGetReturnValues(
        fixture.publicStaking,
        "getPosition",
        users[i] as SignerWithAddress,
        [tokensID[i]]
      );
      // compute payout
      let diffAccum =
        tokenStateAccum - userPosition.accumulatorToken.toBigInt();
      let payoutEst = diffAccum * sharesPerUser;
      if (totalShares == userPosition.shares.toBigInt()) {
        payoutEst += tokenStateSlush;
        tokenStateSlush = 0n;
      }
      let payoutRem = payoutEst;
      payoutEst /= scaleFactor;
      payoutRem -= payoutEst * scaleFactor;
      tokenStateSlush += payoutRem;

      await collectTokensCheckAndUpdateState(
        fixture.publicStaking,
        fixture.aToken,
        payoutEst,
        i,
        users,
        tokensID,
        expectedState,
        "After collect 5-" + i,
        tokenStateSlush
      );
    }
    await mineBlocks(2n);

    // deposit eth this time
    let ethStateAccum = 0n;
    let ethStateSlush = 0n;
    amountDeposited = ethers.utils.parseEther("1000").toBigInt();
    ethStateSlush += 1000n * 10n ** 18n * scaleFactor;
    await depositEthCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 6 Eth"
    );

    // Perform slushSkim
    totalSharesBN = await fixture.publicStaking.getTotalShares();
    totalShares = totalSharesBN.toBigInt();
    deltaAccum = tokenStateSlush / totalShares;
    tokenStateSlush -= deltaAccum * totalShares;
    tokenStateAccum += deltaAccum;
    let deltaAccumEth = ethStateSlush / totalShares;
    ethStateSlush -= deltaAccumEth * totalShares;
    ethStateAccum += deltaAccumEth;
    // get position info
    [userPosition] = await callFunctionAndGetReturnValues(
      fixture.publicStaking,
      "getPosition",
      users[0] as SignerWithAddress,
      [tokensID[0]]
    );
    // compute payout token
    diffAccum = tokenStateAccum - userPosition.accumulatorToken.toBigInt();
    payoutEst = diffAccum * sharesPerUser;
    if (totalShares == userPosition.shares.toBigInt()) {
      payoutEst += tokenStateSlush;
      tokenStateSlush = 0n;
    }
    payoutRem = payoutEst;
    payoutEst /= scaleFactor;
    payoutRem -= payoutEst * scaleFactor;
    tokenStateSlush += payoutRem;
    payoutEst += sharesPerUser; // include because burning position
    // compute payout eth
    let diffAccumEth = ethStateAccum - userPosition.accumulatorEth.toBigInt();
    let payoutEstEth = diffAccumEth * sharesPerUser;
    if (totalShares == userPosition.shares.toBigInt()) {
      payoutEstEth += ethStateSlush;
      ethStateSlush = 0n;
    }
    let payoutRemEth = payoutEstEth;
    payoutEstEth /= scaleFactor;
    payoutRemEth -= payoutEstEth * scaleFactor;
    ethStateSlush += payoutRemEth;

    // Start to burn the positions
    await burnPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      sharesPerUser,
      payoutEstEth,
      payoutEst,
      0,
      users,
      tokensID,
      expectedState,
      "After burn 1",
      ethStateSlush,
      tokenStateSlush
    );

    // Perform slushSkim
    totalSharesBN = await fixture.publicStaking.getTotalShares();
    totalShares = totalSharesBN.toBigInt();
    deltaAccum = tokenStateSlush / totalShares;
    tokenStateSlush -= deltaAccum * totalShares;
    tokenStateAccum += deltaAccum;
    deltaAccumEth = ethStateSlush / totalShares;
    ethStateSlush -= deltaAccumEth * totalShares;
    ethStateAccum += deltaAccumEth;
    // get position info
    [userPosition] = await callFunctionAndGetReturnValues(
      fixture.publicStaking,
      "getPosition",
      users[1] as SignerWithAddress,
      [tokensID[1]]
    );
    // compute payout token
    diffAccum = tokenStateAccum - userPosition.accumulatorToken.toBigInt();
    payoutEst = diffAccum * sharesPerUser;
    if (totalShares == userPosition.shares.toBigInt()) {
      payoutEst += tokenStateSlush;
      tokenStateSlush = 0n;
    }
    payoutRem = payoutEst;
    payoutEst /= scaleFactor;
    payoutRem -= payoutEst * scaleFactor;
    tokenStateSlush += payoutRem;
    payoutEst += sharesPerUser; // include because burning position
    // compute payout eth
    diffAccumEth = ethStateAccum - userPosition.accumulatorEth.toBigInt();
    payoutEstEth = diffAccumEth * sharesPerUser;
    if (totalShares == userPosition.shares.toBigInt()) {
      payoutEstEth += ethStateSlush;
      ethStateSlush = 0n;
    }
    payoutRemEth = payoutEstEth;
    payoutEstEth /= scaleFactor;
    payoutRemEth -= payoutEstEth * scaleFactor;
    ethStateSlush += payoutRemEth;

    await burnPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      sharesPerUser,
      payoutEstEth,
      payoutEst,
      1,
      users,
      tokensID,
      expectedState,
      "After burn 2",
      ethStateSlush,
      tokenStateSlush
    );

    // Perform slushSkim
    totalSharesBN = await fixture.publicStaking.getTotalShares();
    totalShares = totalSharesBN.toBigInt();
    deltaAccum = tokenStateSlush / totalShares;
    tokenStateSlush -= deltaAccum * totalShares;
    tokenStateAccum += deltaAccum;
    deltaAccumEth = ethStateSlush / totalShares;
    ethStateSlush -= deltaAccumEth * totalShares;
    ethStateAccum += deltaAccumEth;
    // get position info
    [userPosition] = await callFunctionAndGetReturnValues(
      fixture.publicStaking,
      "getPosition",
      users[2] as SignerWithAddress,
      [tokensID[2]]
    );
    // compute payout token
    diffAccum = tokenStateAccum - userPosition.accumulatorToken.toBigInt();
    payoutEst = diffAccum * sharesPerUser;
    if (totalShares == userPosition.shares.toBigInt()) {
      payoutEst += tokenStateSlush;
      tokenStateSlush = 0n;
    }
    payoutRem = payoutEst;
    payoutEst /= scaleFactor;
    payoutRem -= payoutEst * scaleFactor;
    tokenStateSlush += payoutRem;
    payoutEst += sharesPerUser; // include because burning position
    // compute payout eth
    diffAccumEth = ethStateAccum - userPosition.accumulatorEth.toBigInt();
    payoutEstEth = diffAccumEth * sharesPerUser;
    if (totalShares == userPosition.shares.toBigInt()) {
      payoutEstEth += ethStateSlush;
      ethStateSlush = 0n;
    }
    payoutRemEth = payoutEstEth;
    payoutEstEth /= scaleFactor;
    payoutRemEth -= payoutEstEth * scaleFactor;
    ethStateSlush += payoutRemEth;

    console.log("Before Burn 3");
    // last user should get the shares + slushes since he's the last user exiting the staking contract
    await burnPositionCheckAndUpdateState(
      fixture.publicStaking,
      fixture.aToken,
      sharesPerUser,
      payoutEstEth,
      payoutEst,
      2,
      users,
      tokensID,
      expectedState,
      "After burn 3",
      ethStateSlush,
      tokenStateSlush
    );
  });
});
