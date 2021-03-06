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

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    await fixture.aToken.approve(
      fixture.publicStaking.address,
      ethers.utils.parseUnits("100000", 18)
    );
    users = await createUsers(numberUsers);
    const baseAmount = ethers.utils.parseUnits("10000", 1).toBigInt();
    for (let i = 0; i < numberUsers; i++) {
      await fixture.aToken.transfer(await users[i].getAddress(), baseAmount);
      await fixture.aToken
        .connect(users[i])
        .approve(fixture.publicStaking.address, baseAmount);
    }
  });

  it("Slush flush into accumulator", async function () {
    const sharesPerUser = 10n;
    const tokensID: number[] = [];
    for (let i = 0; i < numberUsers; i++) {
      const tx = await fixture.publicStaking
        .connect(users[i])
        .mint(sharesPerUser);
      tokensID.push(await getTokenIdFromTx(tx));
    }
    let credits = 0n;
    let debits = 0n;
    for (let j = 0; j < 2; j++) {
      // expected 1000//3 = 333. So each validator should receive 333 and we should have +1 in the slush
      // per deposit
      await fixture.publicStaking.depositEth(42, {
        value: 1000,
      });
      credits += 1000n;
      for (let i = 0; i < numberUsers; i++) {
        const [payout] = await callFunctionAndGetReturnValues(
          fixture.publicStaking,
          "collectEth",
          users[i] as SignerWithAddress,
          [tokensID[i]]
        );
        expect(payout.toBigInt()).to.be.equals(
          333n,
          `User ${i} didn't collect the expected amount!`
        );
        debits += payout.toBigInt();
      }
    }
    let [, slush] = await fixture.publicStaking.getEthAccumulator();
    const scaleFactor = (
      await fixture.publicStaking.getAccumulatorScaleFactor()
    ).toBigInt();
    let expectedAmount = credits - debits;
    // The value in the slush is scaled by the scale factor
    expect(slush.toBigInt()).to.be.equals(
      expectedAmount * scaleFactor,
      "Slush does not correspond to expected value!"
    );
    expect(slush.toBigInt()).to.be.equals(
      2n * scaleFactor,
      "Slush does not correspond to expected value 2!"
    );
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      30n,
      expectedAmount
    );
    // with a final deposit of 2000 we should end up with 2002 since the accumulator contains 2
    // 2002//3 = 667 for each validator + 1 in the accumulator as remainder
    await fixture.publicStaking.depositEth(42, {
      value: 2000,
    });
    credits += 2000n;
    expectedAmount = credits - debits;
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      30n,
      expectedAmount
    );
    for (let i = 0; i < numberUsers; i++) {
      const [payout] = await callFunctionAndGetReturnValues(
        fixture.publicStaking,
        "collectEth",
        users[i] as SignerWithAddress,
        [tokensID[i]]
      );
      expect(payout.toBigInt()).to.be.equals(
        667n,
        `User ${i} didn't collect the expected amount!`
      );
      debits += payout.toBigInt();
    }
    expectedAmount = credits - debits;
    [, slush] = await fixture.publicStaking.getEthAccumulator();
    expect(slush.toBigInt()).to.be.equals(
      expectedAmount * scaleFactor,
      "Slush does not correspond to expected value after 2nd deposit!"
    );
    expect(slush.toBigInt()).to.be.equals(
      1n * scaleFactor,
      "Slush does not correspond to expected value after 2nd deposit 2!"
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
    let credits = 0n;
    let debits = 0n;
    for (let j = 0; j < 37; j++) {
      await fixture.publicStaking.depositEth(42, {
        value: 7,
      });
      credits += 7n;
      for (let i = 0; i < numberUsers; i++) {
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
      expectedAmount * scaleFactor,
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
      credits += constCreditValues[j];
      expectedAmount = credits - debits;
      await assertTotalReserveAndZeroExcess(
        fixture.publicStaking,
        totalMinted,
        expectedAmount
      );
      for (let i = 0; i < numberUsers; i++) {
        debits += (
          await collectEth(fixture.publicStaking, users[i], tokensID[i])
        ).toBigInt();
      }
      expectedAmount = credits - debits;
      [, slush] = await fixture.publicStaking.getEthAccumulator();
      expect(slush.toBigInt()).to.be.equals(
        expectedAmount * scaleFactor,
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
    await fixture.publicStaking.depositEth(42, {
      value: depositAmount,
    });
    credits += depositAmount;
    debits += (
      await collectEth(fixture.publicStaking, users[0], tokensID[0])
    ).toBigInt();
    debits += (
      await collectEth(fixture.publicStaking, users[1], tokensID[1])
    ).toBigInt();
    depositAmount = 11_209873167895423687n;
    await fixture.publicStaking.depositEth(42, {
      value: depositAmount,
    });
    credits += depositAmount;
    debits += (
      await collectEth(fixture.publicStaking, users[0], tokensID[0])
    ).toBigInt();
    depositAmount = 156_209873167895423687n;
    await fixture.publicStaking.depositEth(42, {
      value: depositAmount,
    });
    credits += depositAmount;
    for (let i = 0; i < numberUsers; i++) {
      debits += (
        await collectEth(fixture.publicStaking, users[i], tokensID[i])
      ).toBigInt();
    }
    expectedAmount = credits - debits;
    [, slush] = await fixture.publicStaking.getEthAccumulator();
    expect(slush.toBigInt()).to.be.equals(
      expectedAmount * scaleFactor,
      `Slush does not correspond to expected value after all deposits!`
    );
    await assertTotalReserveAndZeroExcess(
      fixture.publicStaking,
      totalMinted,
      expectedAmount
    );
  });
});
