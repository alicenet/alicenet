import { loadFixture } from "@nomicfoundation/hardhat-network-helpers";
import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import {
  BaseTokensFixture,
  callFunctionAndGetReturnValues,
  createUsers,
  getBaseTokensFixture,
  getTokenIdFromTx,
} from "../../setup";
import { assertTotalReserveAndZeroExcess, collectEth } from "../setup";

describe("PublicStaking: Accumulator and slush invariance", async () => {
  let fixture: BaseTokensFixture;
  let users: SignerWithAddress[];
  const numberUsers = 3;

  async function deployFixture() {
    const fixture = await getBaseTokensFixture();
    await fixture.aToken.approve(
      fixture.publicStaking.address,
      ethers.utils.parseUnits("100000", 18)
    );
    const users = await createUsers(numberUsers);
    const baseAmount = ethers.utils.parseUnits("10000", 1);
    for (let i = 0; i < numberUsers; i++) {
      await fixture.aToken.transfer(await users[i].getAddress(), baseAmount);
      await fixture.aToken
        .connect(users[i])
        .approve(fixture.publicStaking.address, baseAmount);
    }
    return { fixture, users };
  }

  beforeEach(async function () {
    ({ fixture, users } = await loadFixture(deployFixture));
  });

  it("Slush flush into accumulator", async function () {
    const scaleFactor = (
      await fixture.publicStaking.getAccumulatorScaleFactor()
    ).toBigInt();
    const sharesPerUser = 10n;
    const tokensID: number[] = [];
    for (let i = 0; i < numberUsers; i++) {
      const tx = await fixture.publicStaking
        .connect(users[i])
        .mint(sharesPerUser);
      tokensID.push(await getTokenIdFromTx(tx));
    }
    const totalSharesBN = await fixture.publicStaking.getTotalReserveAToken();
    const totalShares = totalSharesBN.toBigInt();
    let ethStateAccum = 0n;
    let ethStateSlush = 0n;
    let credits = 0n;
    let debits = 0n;
    for (let j = 0; j < 2; j++) {
      await fixture.publicStaking.depositEth(42, {
        value: 1000,
      });
      ethStateSlush += 1000n * scaleFactor;
      credits += 1000n;
      for (let i = 0; i < numberUsers; i++) {
        // Perform slushSkim
        const deltaAccum = ethStateSlush / totalShares;
        ethStateSlush -= deltaAccum * totalShares;
        ethStateAccum += deltaAccum;
        // compute payout
        const [userPosition] = await callFunctionAndGetReturnValues(
          fixture.publicStaking,
          "getPosition",
          users[i] as SignerWithAddress,
          [tokensID[i]]
        );
        // compute payout
        const diffAccum =
          ethStateAccum - userPosition.accumulatorEth.toBigInt();
        let payoutEst = diffAccum * sharesPerUser;
        let payoutRem = payoutEst;
        payoutEst /= scaleFactor;
        payoutRem -= payoutEst * scaleFactor;
        ethStateSlush += payoutRem;

        const [payout] = await callFunctionAndGetReturnValues(
          fixture.publicStaking,
          "collectEth",
          users[i] as SignerWithAddress,
          [tokensID[i]]
        );

        expect(payout.toBigInt()).to.be.equals(
          payoutEst,
          `User ${i} didn't collect the expected amount!`
        );
        debits += payout.toBigInt();
      }
    }
    let [, slush] = await fixture.publicStaking.getEthAccumulator();
    let expectedAmount = credits - debits;
    // The value in the slush is scaled by the scale factor
    expect(slush.toBigInt()).to.be.equals(
      ethStateSlush,
      "Slush does not correspond to expected value!"
    );
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      30n,
      expectedAmount
    );
    await fixture.publicStaking.depositEth(42, {
      value: 2000,
    });
    ethStateSlush += 2000n * scaleFactor;
    credits += 2000n;
    expectedAmount = credits - debits;
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      30n,
      expectedAmount
    );
    for (let i = 0; i < numberUsers; i++) {
      // Perform slushSkim
      const deltaAccum = ethStateSlush / totalShares;
      ethStateSlush -= deltaAccum * totalShares;
      ethStateAccum += deltaAccum;
      // compute payout
      const [userPosition] = await callFunctionAndGetReturnValues(
        fixture.publicStaking,
        "getPosition",
        users[i] as SignerWithAddress,
        [tokensID[i]]
      );
      // compute payout
      const diffAccum = ethStateAccum - userPosition.accumulatorEth.toBigInt();
      let payoutEst = diffAccum * sharesPerUser;
      let payoutRem = payoutEst;
      payoutEst /= scaleFactor;
      payoutRem -= payoutEst * scaleFactor;
      ethStateSlush += payoutRem;

      const [payout] = await callFunctionAndGetReturnValues(
        fixture.publicStaking,
        "collectEth",
        users[i] as SignerWithAddress,
        [tokensID[i]]
      );
      expect(payout.toBigInt()).to.be.equals(
        payoutEst,
        `User ${i} didn't collect the expected amount!`
      );
      debits += payout.toBigInt();
    }
    expectedAmount = credits - debits;
    [, slush] = await fixture.publicStaking.getEthAccumulator();
    expect(slush.toBigInt()).to.be.equals(
      ethStateSlush,
      "Slush does not correspond to expected value after 2nd deposit!"
    );
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      30n,
      expectedAmount
    );
  });

  it("Slush invariance", async function () {
    const tokensID: number[] = [];
    const mintValues = [3333n, 111n, 7n];
    const totalMinted = 3333n + 111n + 7n;
    const scaleFactor = (
      await fixture.publicStaking.getAccumulatorScaleFactor()
    ).toBigInt();
    let ethStateAccum = 0n;
    let ethStateSlush = 0n;

    for (let i = 0; i < numberUsers; i++) {
      const tx = await fixture.publicStaking
        .connect(users[i])
        .mint(mintValues[i]);
      tokensID.push(await getTokenIdFromTx(tx));
    }
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      totalMinted,
      0n
    );

    const totalSharesBN = await fixture.publicStaking.getTotalShares();
    const totalShares = totalSharesBN.toBigInt();

    let credits = 0n;
    let debits = 0n;
    for (let j = 0; j < 37; j++) {
      await fixture.publicStaking.depositEth(42, {
        value: 7,
      });
      credits += 7n;
      ethStateSlush += 7n * scaleFactor;
      for (let i = 0; i < numberUsers; i++) {
        // Prep for collect
        const deltaAccum = ethStateSlush / totalShares;
        ethStateSlush -= deltaAccum * totalShares;
        ethStateAccum += deltaAccum;
        // get position info
        const [userPosition] = await callFunctionAndGetReturnValues(
          fixture.publicStaking,
          "getPosition",
          users[i] as SignerWithAddress,
          [tokensID[i]]
        );
        // compute payout eth
        const diffAccum =
          ethStateAccum - userPosition.accumulatorEth.toBigInt();
        let payoutEst = diffAccum * userPosition.shares.toBigInt();
        if (totalShares === userPosition.shares.toBigInt()) {
          payoutEst += ethStateSlush;
          ethStateSlush = 0n;
        }
        let payoutRem = payoutEst;
        payoutEst /= scaleFactor;
        payoutRem -= payoutEst * scaleFactor;
        ethStateSlush += payoutRem;

        debits += (
          await collectEth(fixture.publicStaking, users[i], tokensID[i])
        ).toBigInt();
      }
    }

    let [, slush] = await fixture.publicStaking.getEthAccumulator();
    let expectedAmount = credits - debits;
    // The value in the slush is scaled by the scale factor
    // As long as all the users have withdrawal their dividends this should hold true
    expect(slush.toBigInt()).to.be.equals(
      ethStateSlush,
      "Slush does not correspond to expected value!"
    );
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      totalMinted,
      expectedAmount
    );

    const constCreditValues = [credits, 13457811n];
    for (let j = 0; j < constCreditValues.length; j++) {
      await fixture.publicStaking.depositEth(42, {
        value: constCreditValues[j],
      });
      ethStateSlush += constCreditValues[j] * scaleFactor;

      credits += constCreditValues[j];
      let expectedAmount = credits - debits;
      await assertTotalReserveAndZeroExcess(
        fixture.publicStaking,
        totalMinted,
        expectedAmount
      );

      for (let i = 0; i < numberUsers; i++) {
        // Prep for collect
        const deltaAccum = ethStateSlush / totalShares;
        ethStateSlush -= deltaAccum * totalShares;
        ethStateAccum += deltaAccum;
        // get position info
        const [userPosition] = await callFunctionAndGetReturnValues(
          fixture.publicStaking,
          "getPosition",
          users[i] as SignerWithAddress,
          [tokensID[i]]
        );
        // compute payout eth
        const diffAccum =
          ethStateAccum - userPosition.accumulatorEth.toBigInt();
        let payoutEst = diffAccum * userPosition.shares.toBigInt();
        if (totalShares === userPosition.shares.toBigInt()) {
          payoutEst += ethStateSlush;
          ethStateSlush = 0n;
        }
        let payoutRem = payoutEst;
        payoutEst /= scaleFactor;
        payoutRem -= payoutEst * scaleFactor;
        ethStateSlush += payoutRem;

        debits += (
          await collectEth(fixture.publicStaking, users[i], tokensID[i])
        ).toBigInt();
      }

      expectedAmount = credits - debits;
      [, slush] = await fixture.publicStaking.getEthAccumulator();
      expect(slush.toBigInt()).to.be.equals(
        ethStateSlush,
        `Slush does not correspond to expected value after ${j + 1} deposit!`
      );

      await assertTotalReserveAndZeroExcess(
        fixture.publicStaking,
        totalMinted,
        expectedAmount
      );
    }

    // randomly deposit and only some of the validators collect
    let depositAmount = 1111_209873167895423687n;
    ethStateSlush += 1111_209873167895423687n * scaleFactor;
    await fixture.publicStaking.depositEth(42, {
      value: depositAmount,
    });
    credits += depositAmount;

    // Prep for collect
    let deltaAccum = ethStateSlush / totalShares;
    ethStateSlush -= deltaAccum * totalShares;
    ethStateAccum += deltaAccum;
    // get position info
    let [userPosition] = await callFunctionAndGetReturnValues(
      fixture.publicStaking,
      "getPosition",
      users[0] as SignerWithAddress,
      [tokensID[0]]
    );
    // compute payout eth
    let diffAccum = ethStateAccum - userPosition.accumulatorEth.toBigInt();
    let payoutEst = diffAccum * userPosition.shares.toBigInt();
    if (totalShares === userPosition.shares.toBigInt()) {
      payoutEst += ethStateSlush;
      ethStateSlush = 0n;
    }
    let payoutRem = payoutEst;
    payoutEst /= scaleFactor;
    payoutRem -= payoutEst * scaleFactor;
    ethStateSlush += payoutRem;

    debits += (
      await collectEth(fixture.publicStaking, users[0], tokensID[0])
    ).toBigInt();

    // Prep for collect
    deltaAccum = ethStateSlush / totalShares;
    ethStateSlush -= deltaAccum * totalShares;
    ethStateAccum += deltaAccum;
    // get position info
    [userPosition] = await callFunctionAndGetReturnValues(
      fixture.publicStaking,
      "getPosition",
      users[1] as SignerWithAddress,
      [tokensID[1]]
    );
    // compute payout eth
    diffAccum = ethStateAccum - userPosition.accumulatorEth.toBigInt();
    payoutEst = diffAccum * userPosition.shares.toBigInt();
    if (totalShares === userPosition.shares.toBigInt()) {
      payoutEst += ethStateSlush;
      ethStateSlush = 0n;
    }
    payoutRem = payoutEst;
    payoutEst /= scaleFactor;
    payoutRem -= payoutEst * scaleFactor;
    ethStateSlush += payoutRem;

    debits += (
      await collectEth(fixture.publicStaking, users[1], tokensID[1])
    ).toBigInt();
    depositAmount = 11_209873167895423687n;
    ethStateSlush += 11_209873167895423687n * scaleFactor;
    await fixture.publicStaking.depositEth(42, {
      value: depositAmount,
    });
    credits += depositAmount;

    // Prep for collect
    deltaAccum = ethStateSlush / totalShares;
    ethStateSlush -= deltaAccum * totalShares;
    ethStateAccum += deltaAccum;
    // get position info
    [userPosition] = await callFunctionAndGetReturnValues(
      fixture.publicStaking,
      "getPosition",
      users[0] as SignerWithAddress,
      [tokensID[0]]
    );
    // compute payout eth
    diffAccum = ethStateAccum - userPosition.accumulatorEth.toBigInt();
    payoutEst = diffAccum * userPosition.shares.toBigInt();
    if (totalShares === userPosition.shares.toBigInt()) {
      payoutEst += ethStateSlush;
      ethStateSlush = 0n;
    }
    payoutRem = payoutEst;
    payoutEst /= scaleFactor;
    payoutRem -= payoutEst * scaleFactor;
    ethStateSlush += payoutRem;

    debits += (
      await collectEth(fixture.publicStaking, users[0], tokensID[0])
    ).toBigInt();
    depositAmount = 156_209873167895423687n;
    ethStateSlush += 156_209873167895423687n * scaleFactor;
    await fixture.publicStaking.depositEth(42, {
      value: depositAmount,
    });
    credits += depositAmount;
    for (let i = 0; i < numberUsers; i++) {
      // Prep for collect
      const deltaAccum = ethStateSlush / totalShares;
      ethStateSlush -= deltaAccum * totalShares;
      ethStateAccum += deltaAccum;
      // get position info
      const [userPosition] = await callFunctionAndGetReturnValues(
        fixture.publicStaking,
        "getPosition",
        users[i] as SignerWithAddress,
        [tokensID[i]]
      );
      // compute payout eth
      const diffAccum = ethStateAccum - userPosition.accumulatorEth.toBigInt();
      let payoutEst = diffAccum * userPosition.shares.toBigInt();
      if (totalShares === userPosition.shares.toBigInt()) {
        payoutEst += ethStateSlush;
        ethStateSlush = 0n;
      }
      let payoutRem = payoutEst;
      payoutEst /= scaleFactor;
      payoutRem -= payoutEst * scaleFactor;
      ethStateSlush += payoutRem;

      debits += (
        await collectEth(fixture.publicStaking, users[i], tokensID[i])
      ).toBigInt();
    }
    expectedAmount = credits - debits;
    [, slush] = await fixture.publicStaking.getEthAccumulator();
    expect(slush.toBigInt()).to.be.equals(
      ethStateSlush,
      `Slush does not correspond to expected value after all deposits!`
    );
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      totalMinted,
      expectedAmount
    );
  });
});
