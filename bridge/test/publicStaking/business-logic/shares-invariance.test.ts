import { SignerWithAddress } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import {
  BaseTokensFixture,
  createUsers,
  getBaseTokensFixture,
  mineBlocks,
} from "../../setup";
import { assertTotalReserveAndZeroExcess } from "../setup";

describe("PublicStaking: Shares Invariance", async () => {
  let fixture: BaseTokensFixture;
  let users: SignerWithAddress[];
  const numberUsers = 50;

  beforeEach(async function () {
    fixture = await getBaseTokensFixture();
    await fixture.madToken.approve(
      fixture.publicStaking.address,
      ethers.utils.parseUnits("100000", 18)
    );
    users = await createUsers(numberUsers);
  });

  it("Invariance should hold with multiple users", async function () {
    const baseAmount = ethers.utils.parseUnits("1000000", 18).toBigInt();
    let shares = 0n;
    for (let i = 0; i < numberUsers; i++) {
      const userAmount = baseAmount + BigInt(i);
      await fixture.madToken.transfer(users[i].address, userAmount);
      await fixture.madToken
        .connect(users[i])
        .approve(fixture.publicStaking.address, userAmount);
      await fixture.publicStaking.connect(users[i]).mint(userAmount);
      shares += userAmount;
    }
    expect(
      (await fixture.publicStaking.getTotalShares()).toBigInt()
    ).to.be.equals(shares);
    await assertTotalReserveAndZeroExcess(fixture.publicStaking, shares, 0n);
    await mineBlocks(2n);
    for (let i = 0; i < numberUsers; i++) {
      if (i % 3 === 0) {
        const userAmount = baseAmount + BigInt(i);
        await fixture.publicStaking.connect(users[i]).burn(i + 1);
        shares -= userAmount;
      }
    }
    expect(
      (await fixture.publicStaking.getTotalShares()).toBigInt()
    ).to.be.equals(shares);
    await assertTotalReserveAndZeroExcess(fixture.publicStaking, shares, 0n);
  });
});
