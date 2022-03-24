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
  assertAccumulatorAndSlushEth,
  assertAccumulatorAndSlushToken,
  assertERC20Balance,
  assertPositions,
  assertTotalReserveAndZeroExcess,
  estimateAndCollectTokens,
  mintPosition,
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
    const baseAmount = ethers.utils.parseUnits("10000", 1).toBigInt();
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
    const scaleFactor = (
      await fixture.publicStaking.getAccumulatorScaleFactor()
    ).toBigInt();
    const tokensID: number[] = [];
    let [tokenID, expectedPosition] = await mintPosition(
      fixture.publicStaking,
      users[0],
      sharesPerUser
    );
    tokensID.push(tokenID);

    await assertAccumulatorAndSlushEth(fixture.publicStaking, 0n, 0n);
    await assertAccumulatorAndSlushToken(fixture.publicStaking, 0n, 0n);

    let amountDeposited = 50n;
    await fixture.publicStaking.depositToken(42, amountDeposited);
    await assertTotalReserveAndZeroExcess(fixture.publicStaking, 150n, 0n);

    await assertAccumulatorAndSlushEth(fixture.publicStaking, 0n, 0n);
    let accTokenExpected = (amountDeposited * scaleFactor) / sharesPerUser;
    await assertAccumulatorAndSlushToken(
      fixture.publicStaking,
      accTokenExpected,
      0n
    );
    const [accumulator, slush] =
      await fixture.publicStaking.getTokenAccumulator();

    expect(
      slush.toBigInt() + accumulator.toBigInt() * sharesPerUser
    ).to.be.equals(
      amountDeposited * scaleFactor,
      "Accumulator and slush expected value dont match!"
    );

    let expectedTokensAmount = 150n;
    await assertERC20Balance(
      fixture.madToken,
      fixture.publicStaking.address,
      expectedTokensAmount
    );

    await estimateAndCollectTokens(
      fixture.publicStaking,
      tokensID[0],
      users[0],
      50n
    );

    expectedTokensAmount = expectedTokensAmount - 50n;
    await assertERC20Balance(
      fixture.madToken,
      fixture.publicStaking.address,
      expectedTokensAmount
    );

    await assertTotalReserveAndZeroExcess(fixture.publicStaking, 100n, 0n);

    await assertAccumulatorAndSlushToken(
      fixture.publicStaking,
      accTokenExpected,
      0n
    );
    [tokenID, expectedPosition] = await mintPosition(
      fixture.publicStaking,
      users[1],
      sharesPerUser,
      0n,
      accTokenExpected
    );
    tokensID.push(tokenID);
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      sharesPerUser * 2n,
      0n
    );

    await assertAccumulatorAndSlushToken(
      fixture.publicStaking,
      accTokenExpected,
      0n
    );

    await assertAccumulatorAndSlushEth(fixture.publicStaking, 0n, 0n);
    amountDeposited = 500n;
    expectedTokensAmount = sharesPerUser * 2n + amountDeposited;
    await fixture.publicStaking.depositToken(42, amountDeposited);

    accTokenExpected += (amountDeposited * scaleFactor) / (2n * sharesPerUser);
    await assertAccumulatorAndSlushToken(
      fixture.publicStaking,
      accTokenExpected,
      0n
    );

    await assertAccumulatorAndSlushEth(fixture.publicStaking, 0n, 0n);
    await assertERC20Balance(
      fixture.madToken,
      fixture.publicStaking.address,
      expectedTokensAmount
    );
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      expectedTokensAmount,
      0n
    );
    await estimateAndCollectTokens(
      fixture.publicStaking,
      tokensID[0],
      users[0],
      amountDeposited / 2n
    );
    expectedTokensAmount -= amountDeposited / 2n;
    await assertERC20Balance(
      fixture.madToken,
      fixture.publicStaking.address,
      expectedTokensAmount
    );
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      expectedTokensAmount,
      0n
    );
    await assertAccumulatorAndSlushToken(
      fixture.publicStaking,
      accTokenExpected,
      0n
    );

    await assertAccumulatorAndSlushEth(fixture.publicStaking, 0n, 0n);
    expectedPosition.accumulatorToken = accTokenExpected;
    await assertPositions(fixture.publicStaking, tokensID[0], expectedPosition);

    await estimateAndCollectTokens(
      fixture.publicStaking,
      tokensID[1],
      users[1],
      amountDeposited / 2n
    );

    expectedTokensAmount -= amountDeposited / 2n;
    await assertERC20Balance(
      fixture.madToken,
      fixture.publicStaking.address,
      expectedTokensAmount
    );
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      expectedTokensAmount,
      0n
    );
    await assertAccumulatorAndSlushToken(
      fixture.publicStaking,
      accTokenExpected,
      0n
    );

    await assertAccumulatorAndSlushEth(fixture.publicStaking, 0n, 0n);
    expectedPosition.accumulatorToken = accTokenExpected;
    await assertPositions(fixture.publicStaking, tokensID[1], expectedPosition);
  });
});
