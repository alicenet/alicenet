import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { ethers } from "hardhat";
import { ALCA, HugeAccumulatorStaking } from "../../../typechain-types";
import {
  createUsers,
  deployAliceNetFactory,
  deployUpgradeableWithFactory,
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
  let alca: ALCA;
  const numberUsers: number = 2;
  let users: SignerWithAddress[] = [];

  async function deployFixture() {
    const [adminSigner] = await ethers.getSigners();
    await preFixtureSetup();
    const legacyToken = await (
      await ethers.getContractFactory("LegacyToken")
    ).deploy();
    const factory = await deployAliceNetFactory(
      adminSigner,
      legacyToken.address
    );

    const alca = await ethers.getContractAt(
      "ALCA",
      await factory.lookup(ethers.utils.formatBytes32String("ALCA"))
    );
    const stakingContract = (await deployUpgradeableWithFactory(
      factory,
      "HugeAccumulatorStaking",
      "PublicStaking"
    )) as HugeAccumulatorStaking;
    await posFixtureSetup(factory, alca);

    await alca.approve(
      stakingContract.address,
      ethers.utils.parseUnits("100000", 18)
    );
    const users = await createUsers(numberUsers);
    const baseAmount = ethers.utils.parseUnits("100", 0).toBigInt();
    for (let i = 0; i < numberUsers; i++) {
      await alca.transfer(await users[i].getAddress(), baseAmount);
      await alca.connect(users[i]).approve(stakingContract.address, baseAmount);
    }
    await mineBlocks(2n);

    return { factory, stakingContract, alca, users };
  }

  beforeEach(async function () {
    ({ stakingContract, alca, users } = await loadFixture(deployFixture));
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
      alca,
      users,
      tokensID
    );

    const userMintedAmount = 50n;
    await mintPositionCheckAndUpdateState(
      stakingContract,
      alca,
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
      alca,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 - Token"
    );

    // moving the token accumulator closer to the overflow
    await depositEthCheckAndUpdateState(
      stakingContract,
      alca,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 - Eth"
    );

    await mintPositionCheckAndUpdateState(
      stakingContract,
      alca,
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
      alca,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2 - Token"
    );

    await depositEthCheckAndUpdateState(
      stakingContract,
      alca,
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
        alca,
        expectedCollectedAmount[i],
        i,
        users,
        tokensID,
        expectedState,
        "After collect Token 1 - " + i
      );
      await collectEthCheckAndUpdateState(
        stakingContract,
        alca,
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
      alca,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit token 3"
    );
    await depositEthCheckAndUpdateState(
      stakingContract,
      alca,
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
        alca,
        expectedCollectedAmount[i],
        i,
        users,
        tokensID,
        expectedState,
        "After collect token 2 - " + i
      );
      await collectEthCheckAndUpdateState(
        stakingContract,
        alca,
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
      alca,
      users,
      tokensID
    );

    const userMintedAmount = 50n;
    await mintPositionCheckAndUpdateState(
      stakingContract,
      alca,
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
      alca,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 - Token"
    );

    // moving the token accumulator closer to the overflow
    await depositEthCheckAndUpdateState(
      stakingContract,
      alca,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 1 - Eth"
    );

    await mintPositionCheckAndUpdateState(
      stakingContract,
      alca,
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
      alca,
      amountDeposited,
      users,
      tokensID,
      expectedState,
      "After deposit 2 - Token"
    );

    await depositEthCheckAndUpdateState(
      stakingContract,
      alca,
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
        alca,
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
