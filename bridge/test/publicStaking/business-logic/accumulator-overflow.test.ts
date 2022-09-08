import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import {
  AToken,
  HugeAccumulatorStaking,
  LegacyToken,
} from "../../../typechain-types";
import {
  createUsers,
  deployAliceNetFactory,
  deployStaticWithFactory,
  mineBlocks,
  posFixtureSetup,
  preFixtureSetup,
} from "../../setup";
import {
  assertAccumulatorAndSlushEth,
  assertAccumulatorAndSlushToken,
  burnPositionCheckAndUpdateState,
  collectEthCheckAndUpdateState,
  collectTokensCheckAndUpdateState,
  depositEthCheckAndUpdateState,
  depositTokensCheckAndUpdateState,
  getCurrentState,
  mintPositionCheckAndUpdateState,
} from "../setup";

describe("PublicStaking: Accumulator Overflow", async () => {
  let stakingContract: HugeAccumulatorStaking;
  let aToken: AToken;
  const numberUsers: number = 2;
  let users: SignerWithAddress[] = [];

  beforeEach(async function () {
    const [adminSigner] = await ethers.getSigners();
    await preFixtureSetup();
    const factory = await deployAliceNetFactory(adminSigner);

    const legacyToken = (await deployStaticWithFactory(
      factory,
      "LegacyToken"
    )) as LegacyToken;

    // AToken
    aToken = (await deployStaticWithFactory(
      factory,
      "AToken",
      "AToken",
      undefined,
      [legacyToken.address]
    )) as AToken;

    stakingContract = (await deployStaticWithFactory(
      factory,
      "HugeAccumulatorStaking",
      "PublicStaking"
    )) as HugeAccumulatorStaking;

    await posFixtureSetup(factory, aToken);
    await aToken.approve(
      stakingContract.address,
      ethers.utils.parseUnits("100000", 18)
    );
    users = await createUsers(numberUsers);
    const baseAmount = ethers.utils.parseUnits("100", 0).toBigInt();
    for (let i = 0; i < numberUsers; i++) {
      await aToken.transfer(await users[i].getAddress(), baseAmount);
      await aToken
        .connect(users[i])
        .approve(stakingContract.address, baseAmount);
    }
    await mineBlocks(2n);
  });

  it("Collect Tokens and ETH with overflow in the accumulators", async function () {
    const expectedAccumulatorToken =
      2n ** 168n -
      1n -
      (await stakingContract.getOffsetToOverflow()).toBigInt();
    const expectedAccumulatorEth = expectedAccumulatorToken;
    const tokensID: number[] = [];
    for (let i = 0; i < users.length; i++) {
      tokensID.push(0);
    }

    await assertAccumulatorAndSlushToken(
      stakingContract,
      expectedAccumulatorToken,
      0n
    );
    await assertAccumulatorAndSlushEth(
      stakingContract,
      expectedAccumulatorEth,
      0n
    );

    const expectedState = await getCurrentState(
      stakingContract,
      aToken,
      users,
      tokensID
    );

    const userMintedAmount = 50n;
    await mintPositionCheckAndUpdateState(
      stakingContract,
      aToken,
      userMintedAmount,
      0,
      users,
      tokensID,
      expectedState,
      "After mint 1"
    );

    let amountDeposited = 25n;
    // moving the token accumulator closer to the overflow
    await depositTokensCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 - Token"
    );

    // moving the token accumulator closer to the overflow
    await depositEthCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 - Eth"
    );

    await mintPositionCheckAndUpdateState(
      stakingContract,
      aToken,
      userMintedAmount,
      1,
      users,
      tokensID,
      expectedState,
      "After mint 2"
    );

    // the overflow should happen here, As we are incrementing the accumulator by 150 * 10**16. The
    // following function takes into account the overflow when updating expected state
    amountDeposited = 150n;
    await depositTokensCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2 - Token"
    );

    await depositEthCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2 - Eth"
    );

    await mineBlocks(2n);
    let expectedCollectedAmount = [100n, 75n];
    for (let i = 0; i < 2; i++) {
      await collectTokensCheckAndUpdateState(
        stakingContract,
        aToken,
        expectedCollectedAmount[i],
        i,
        users,
        tokensID,
        expectedState,
        "After collect Token 1 - " + i
      );
      await collectEthCheckAndUpdateState(
        stakingContract,
        aToken,
        expectedCollectedAmount[i],
        i,
        users,
        tokensID,
        expectedState,
        "After collect Eth 1 - " + i
      );
    }

    // after overflow has passed normal behavior is expected!
    amountDeposited = 400n;
    await depositTokensCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit token 3"
    );
    await depositEthCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit eth 3"
    );

    await mineBlocks(2n);

    expectedCollectedAmount = [200n, 200n];
    for (let i = 0; i < 2; i++) {
      await collectTokensCheckAndUpdateState(
        stakingContract,
        aToken,
        expectedCollectedAmount[i],
        i,
        users,
        tokensID,
        expectedState,
        "After collect token 2 - " + i
      );
      await collectEthCheckAndUpdateState(
        stakingContract,
        aToken,
        expectedCollectedAmount[i],
        i,
        users,
        tokensID,
        expectedState,
        "After collect Eth 2 - " + i
      );
    }
  });

  it("Burn position after overflow in the accumulators", async function () {
    const expectedAccumulatorToken =
      2n ** 168n -
      1n -
      (await stakingContract.getOffsetToOverflow()).toBigInt();
    const expectedAccumulatorEth = expectedAccumulatorToken;
    const tokensID: number[] = [];
    for (let i = 0; i < users.length; i++) {
      tokensID.push(0);
    }

    await assertAccumulatorAndSlushToken(
      stakingContract,
      expectedAccumulatorToken,
      0n
    );
    await assertAccumulatorAndSlushEth(
      stakingContract,
      expectedAccumulatorEth,
      0n
    );

    const expectedState = await getCurrentState(
      stakingContract,
      aToken,
      users,
      tokensID
    );

    const userMintedAmount = 50n;
    await mintPositionCheckAndUpdateState(
      stakingContract,
      aToken,
      userMintedAmount,
      0,
      users,
      tokensID,
      expectedState,
      "After mint 1"
    );

    let amountDeposited = 25n;
    // moving the token accumulator closer to the overflow
    await depositTokensCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 - Token"
    );

    // moving the token accumulator closer to the overflow
    await depositEthCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 - Eth"
    );

    await mintPositionCheckAndUpdateState(
      stakingContract,
      aToken,
      userMintedAmount,
      1,
      users,
      tokensID,
      expectedState,
      "After mint 2"
    );

    // the overflow should happen here, As we are incrementing the accumulator by 150 * 10**16. The
    // following function takes into account the overflow when updating expected state
    amountDeposited = 150n;
    await depositTokensCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2 - Token"
    );

    await depositEthCheckAndUpdateState(
      stakingContract,
      aToken,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2 - Eth"
    );
    const expectedPayoutAmountEth = [100n, 75n];
    const expectedPayoutAmountToken = [150n, 125n];
    for (let i = 0; i < 2; i++) {
      await burnPositionCheckAndUpdateState(
        stakingContract,
        aToken,
        userMintedAmount,
        expectedPayoutAmountEth[i],
        expectedPayoutAmountToken[i],
        i,
        users,
        tokensID,
        expectedState,
        "After burn-" + i
      );
    }
  });
});
