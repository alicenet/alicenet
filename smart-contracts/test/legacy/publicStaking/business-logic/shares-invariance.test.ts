import { loadFixture, mine } from "@nomicfoundation/hardhat-network-helpers";
import { HardhatEthersSigner } from "@nomiclabs/hardhat-ethers/signers";
import { expect } from "chai";
import { ethers } from "hardhat";
import {
  BaseTokensFixture,
  createUsers,
  getBaseTokensFixture,
} from "../../setup";
import { assertTotalReserveAndZeroExcess } from "../setup";

describe("PublicStaking: Shares Invariance", async () => {
  let fixture: BaseTokensFixture;
  let users: HardhatEthersSigner[];
  const numberUsers = 50;

  async function deployFixture() {
    const fixture = await getBaseTokensFixture();
    await fixture.alca.approve(
      await fixture.publicStaking.getAddress(),
      ethers.parseUnits("100000", 18)
    );
    const users = await createUsers(numberUsers);
    return { fixture, users };
  }

  beforeEach(async function () {
    ({ fixture, users } = await loadFixture(deployFixture));
  });

  it("Invariance should hold with multiple users", async function () {
    const baseAmount = ethers.parseUnits("1000000", 18);
    let shares = 0n;
    for (let i = 0; i < numberUsers; i++) {
      const userAmount = baseAmount + BigInt(i);
      await fixture.alca.transfer(users[i].address, userAmount);
      await fixture.alca
        .connect(users[i])
        .approve(await fixture.publicStaking.getAddress(), userAmount);
      await fixture.publicStaking.connect(users[i]).mint(userAmount);
      shares += userAmount;
    }
    expect(await fixture.publicStaking.getTotalShares()).to.be.equals(shares);
    await assertTotalReserveAndZeroExcess(fixture.publicStaking, shares, 0n);
    await mine(2n);
    for (let i = 0; i < numberUsers; i++) {
      if (i % 3 === 0) {
        const userAmount = baseAmount + BigInt(i);
        await fixture.publicStaking.connect(users[i]).burn(i + 1);
        shares -= userAmount;
      }
    }
    expect(await fixture.publicStaking.getTotalShares()).to.be.equals(shares);
    await assertTotalReserveAndZeroExcess(fixture.publicStaking, shares, 0n);
  });
});
